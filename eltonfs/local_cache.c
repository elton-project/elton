#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/error.h>
#include <elton/local_cache.h>
#include <elton/rpc/server.h>
#include <linux/cred.h>
#include <linux/namei.h>

#define REAL_PATH_MAX 120
#define CACHE_ID_LENGTH 32
#define MAX_CONFLICT_TRIES 30

#define PREFIX_CACHE_DIR "/var/cache/eltonfs"
#define PREFIX_LIB_DIR "/var/lib/eltonfs"

// リモートIDを持つオブジェクトを保存するディレクトリ
#define REMOTE_OBJ_DIR PREFIX_CACHE_DIR "/remote-objects"
// ローカルIDを持つオブジェクトを保存するディレクトリ
#define LOCAL_OBJ_DIR PREFIX_LIB_DIR "/local-objects"

// Create directory to specified path if it is not exist.
static inline int eltonfs_create_dir(const char *pathname) {
  int error = 0;
  struct dentry *dentry = NULL;
  struct path path;

  dentry = kern_path_create(AT_FDCWD, pathname, &path, LOOKUP_DIRECTORY);
  if (IS_ERR(dentry)) {
    if (dentry == ERR_PTR(-EEXIST))
      // Directory already exists.  This error should ignore.
      return 0;
    RETURN_IF(PTR_ERR(dentry));
  }
  GOTO_IF(out, vfs_mkdir(path.dentry->d_inode, dentry, 0700));

out:
  if (dentry)
    done_path_create(&path, dentry);
  return error;
}

int eltonfs_create_cache_dir(void) {
  int error = 0;
  RETURN_IF(eltonfs_create_dir(PREFIX_CACHE_DIR));
  RETURN_IF(eltonfs_create_dir(PREFIX_LIB_DIR));
  RETURN_IF(eltonfs_create_dir(REMOTE_OBJ_DIR));
  RETURN_IF(eltonfs_create_dir(LOCAL_OBJ_DIR));
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

  if (inode->file.cache_inode)
    goto try_open;
  if (inode->file.local_cache_id) {
    struct inode *real_inode =
        eltonfs_get_cache_inode(LOCAL_OBJ_DIR, inode->file.local_cache_id);
    if (IS_ERR(real_inode))
      return ERR_CAST(real_inode);
    inode->file.cache_inode = real_inode;
    goto try_open;
  }
  if (inode->file.object_id) {
    struct inode *real_inode =
        eltonfs_get_cache_inode(LOCAL_OBJ_DIR, inode->file.local_cache_id);
    if (real_inode == ERR_PTR(-ELTON_CACHE_MISS))
      // Should download from remote.
      return NULL;
    if (IS_ERR(real_inode))
      return ERR_CAST(real_inode);
    inode->file.cache_inode = real_inode;
    goto try_open;
  }
  if (file->f_flags & O_CREAT)
    // New file.
    return eltonfs_create_real_file(inode, file);
  return ERR_PTR(-ENOENT);

try_open:
  WARN_ONCE(!inode->file.cache_inode, "cache-inode is null");
  return open_with_fake_path(&file->f_path, file->f_flags,
                             inode->file.cache_inode, current_cred());
}

struct _eltonfs_cache_obj_args {
  const char *oid;
  const struct cred *cred;
};
void *_eltonfs_cache_obj_worker(void *_args) {
  int error = 0;
  const struct _eltonfs_cache_obj_args *args = _args;
  char real_path[REAL_PATH_MAX];
  struct file *real = NULL;
  const struct cred *old_cred;
  struct elton_rpc_ns ns;
  struct get_object_request req = {.id = (char *)args->oid};
  struct get_object_response *res = NULL;
  loff_t offset;

  old_cred = override_creds(args->cred);

  GOTO_IF(out, server.ops->new_session(&server, &ns, NULL));
  GOTO_IF(out, ns.ops->send_struct(&ns, GET_OBJECT_REQUEST_ID, &req));
  GOTO_IF(out, ns.ops->recv_struct(&ns, GET_OBJECT_RESPONSE_ID, (void **)&res));

  eltonfs_cache_fpath_from_cid(real_path, REAL_PATH_MAX, REMOTE_OBJ_DIR,
                               args->oid);
  real = filp_open(real_path, O_CREAT | O_EXCL | O_WRONLY, 0600);
  if (IS_ERR(real))
    GOTO_IF(out, PTR_ERR(real));

  WARN_ONCE(res->body->offset, "cache_obj: invalid offset: offset=%llu",
            res->body->offset);
  // todo: use WRITE_ALL() macro.
  offset = 0;
  GOTO_IF(out, vfs_write(real, res->body->contents, res->body->contents_length,
                         &offset));

out:
  if (res)
    elton_rpc_free_decoded_data(res);
  if (real && !IS_ERR(real))
    filp_close(real, NULL);
  revert_creds(old_cred);
  kfree(args);
  return ERR_PTR(error);
}
int eltonfs_cache_obj_async(struct eltonfs_job *job, const char *oid,
                            struct super_block *sb) {
  struct _eltonfs_cache_obj_args *args = kmalloc(sizeof(*args), GFP_NOFS);
  if (!args)
    return -ENOMEM;
  args->oid = oid;
  args->cred = ((struct eltonfs_info *)sb->s_fs_info)->cred;
  return eltonfs_job_run(job, _eltonfs_cache_obj_worker, args, "get-obj");
}
int eltonfs_cache_obj(const char *oid, struct super_block *sb) {
  int error;
  struct eltonfs_job job;
  void *output;
  error = eltonfs_cache_obj_async(&job, oid, sb);
  if (error)
    return error;
  output = eltonfs_job_wait(&job);
  if (IS_ERR(output))
    return PTR_ERR(output);
  return 0;
}
