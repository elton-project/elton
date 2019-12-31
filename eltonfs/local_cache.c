#include <elton/assert.h>
#include <elton/elton.h>
#include <linux/cred.h>

#define REAL_PATH_MAX 120
#define CACHE_ID_LENGTH 32
#define MAX_CONFLICT_TRIES 30

// リモートIDを持つオブジェクトを保存するディレクトリ
#define REMOTE_OBJ_DIR "/var/cache/eltonfs/remote-objects"
// ローカルIDを持つオブジェクトを保存するディレクトリ
#define LOCAL_OBJ_DIR "/var/lib/eltonfs/local-objects"

static inline int eltonfs_create_dir(const char *path) {
  // https://stackoverflow.com/a/41851045
  struct file *dir;
  dir = filp_open(path, O_DIRECTORY | O_CREAT, 0700);
  if (IS_ERR(dir))
    return PTR_ERR(dir);
  return filp_close(dir, NULL);
}

int eltonfs_create_cache_dir(void) {
  int error = 0;
  RETURN_IF(eltonfs_create_dir(REMOTE_OBJ_DIR "/"));
  RETURN_IF(eltonfs_create_dir(LOCAL_OBJ_DIR "/"));
  return 0;
}

static inline void iget(struct inode *inode) { atomic_inc(&inode->i_count); }

static inline int eltonfs_cache_fpath_from_cid(char *fpath, size_t size,
                                               const char *base_dir,
                                               const char *cid) {
  return snprintf(fpath, size, "%s/%s", base_dir, cid);
}
static inline int eltonfs_cache_fpath_from_int(char *fpath, size_t size,
                                               const char *base_dir, u64 a,
                                               u16 b) {
  return snprintf(fpath, size, "%s/%016llx:%04hx", base_dir, a, b);
}
static inline int eltonfs_cache_id_from_int(char *cache_id, size_t size, u64 a,
                                            u16 b) {
  return snprintf(cache_id, size, "%016llx:%04hx", a, b);
}

// Generate a unique ID.
static int eltonfs_generate_id(const char *base_dir, char fpath[REAL_PATH_MAX],
                               char id[CACHE_ID_LENGTH], struct inode **inode) {
  // seq: Approximative sequential number.
  u64 seq;
  u16 tries;
  struct file *file;
  int len;
  int error;
  BUILD_BUG_ON(CACHE_ID_LENGTH >= REAL_PATH_MAX);

  while (1) {
    seq = (u64)ktime_get_real_fast_ns();
    BUILD_BUG_ON(sizeof(ktime_t) != sizeof(u64));

    for (tries = 0; tries < MAX_CONFLICT_TRIES; tries++) {
      len = eltonfs_cache_fpath_from_int(fpath, REAL_PATH_MAX, base_dir, seq,
                                         tries);
      if (len >= REAL_PATH_MAX)
        return -ENAMETOOLONG;

      file = filp_open(fpath, O_CREAT | O_EXCL, 0700);
      if (file == ERR_PTR(-EEXIST))
        continue;
      if (IS_ERR(file))
        return PTR_ERR(file);

      iget(file->f_inode);
      *inode = file->f_inode;

      error = filp_close(file, NULL);
      if (error)
        return error;

      // Succeed to create a file.  Update id.
      len = eltonfs_cache_id_from_int(id, CACHE_ID_LENGTH, seq, tries);
      BUG_ON(len >= CACHE_ID_LENGTH);
      return 0;
    }
  }
}

// Create local cache file that related to the inode.
// Specified inode MUST NOT associate to cache id or local cache id before this
// function is called.
static struct file *eltonfs_create_real_file(struct eltonfs_inode *inode,
                                             struct file *file) {
  int error;
  char fpath[REAL_PATH_MAX];
  char id[CACHE_ID_LENGTH];
  struct inode *real_inode = NULL;
  unsigned int flags;
  struct file *out = NULL;

  error = eltonfs_generate_id(LOCAL_OBJ_DIR, fpath, id, &real_inode);
  if (error)
    goto err;

  flags = file->f_flags & ~(O_CREAT | O_EXCL);
  out = open_with_fake_path(&file->f_path, flags, real_inode, current_cred());
  if (IS_ERR(out)) {
    error = PTR_ERR(out);
    goto err;
  }

  if (inode->vfs_inode.i_mode & S_IFREG) {
    char *cid = kmalloc(CACHE_ID_LENGTH, GFP_NOFS);
    if (!cid) {
      error = -ENOMEM;
      goto err;
    }
    strncpy(cid, id, CACHE_ID_LENGTH);
    inode->file.local_cache_id = cid;
    iget(real_inode);
    inode->file.cache_inode = real_inode;
  } else {
    DEBUG("create real file with unexpected file type: inode=%px, file=%px",
          inode, file);
    BUG();
  }

  return out;

err:
  if (real_inode)
    iput(real_inode);
  if (out && !IS_ERR(out))
    filp_close(out, NULL);
  return ERR_PTR(error);
}

// Returns an inode associated to cid.
static inline struct inode *eltonfs_get_cache_inode(const char *base_dir,
                                                    const char *cid) {
  int error = 0;
  char fpath[REAL_PATH_MAX];
  struct file *real_file = NULL;
  struct inode *real_inode = NULL;

  eltonfs_cache_fpath_from_cid(fpath, REAL_PATH_MAX, base_dir, cid);
  real_file = filp_open(fpath, O_RDONLY | O_NOATIME, 0);
  if (real_file == ERR_PTR(-ENOENT))
    // Not found
    return NULL;
  if (IS_ERR(real_file))
    return ERR_CAST(real_file);

  iget(real_file->f_inode);
  real_inode = real_file->f_inode;

  error = filp_close(real_file, NULL);
  if (error) {
    iput(real_inode);
    return ERR_PTR(error);
  }
  return real_inode;
}

// Open local cache file that related to the inode.
struct file *eltonfs_open_real_file(struct eltonfs_inode *inode,
                                    struct file *file) {
  // todo: acquire lock
  if (!(inode->vfs_inode.i_mode & S_IFREG)) {
    DEBUG("open real file with unexpected file type: inode=%px, file=%px",
          inode, file);
    BUG();
  }

  if (inode->file.cache_inode) {
    iget(inode->file.cache_inode);
    return open_with_fake_path(&file->f_path, file->f_flags,
                               inode->file.cache_inode, current_cred());
  }
  if (inode->file.local_cache_id) {
    struct file *real_file = NULL;
    struct inode *real_inode =
        eltonfs_get_cache_inode(LOCAL_OBJ_DIR, inode->file.local_cache_id);
    inode->file.cache_inode = real_inode;

    iget(real_inode);
    real_file = open_with_fake_path(&file->f_path, file->f_flags, real_inode,
                                    current_cred());
    if (IS_ERR(real_file))
      iput(real_inode);
    return real_file;
  }
  if (inode->file.object_id)
    // Should download from remote.
    return NULL;
  if (file->f_flags & O_CREAT)
    // New file.
    return eltonfs_create_real_file(inode, file);
  return ERR_PTR(-ENOENT);
}
