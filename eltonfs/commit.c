#include <elton/assert.h>
#include <elton/elton.h>
#include <elton/rpc/server.h>
#include <elton/rpc/struct.h>

// Duplicate NULL terminated string.
static inline int dup_string(char **to, const char *from) {
  int error = 0;
  size_t len = strlen(from);
  char *buff = kmalloc(len + 1, GFP_NOFS);
  if (!buff)
    RETURN_IF(-ENOMEM);
  strcpy(buff, from);
  return 0;
}

static int get_commit_id_by_vid(const char *vid, char **cid) {
  int error = 0;
  struct elton_rpc_ns ns;
  struct notify_latest_commit_request req;
  struct notify_latest_commit *res;

  RETURN_IF(server.ops->new_session(&server, &ns, NULL));
  req.volume_id = vid;
  GOTO_IF(free_ns,
          ns.ops->send_struct(&ns, NOTIFY_LATEST_COMMIT_REQUEST_ID, &req));
  GOTO_IF(free_ns,
          ns.ops->recv_struct(&ns, NOTIFY_LATEST_COMMIT_ID, (void **)&res));

  GOTO_IF(free_res, dup_string(cid, res->commit_id));
free_res:
  elton_rpc_free_decoded_data(res);

free_ns:
  WARN_IF(ns.ops->close(&ns));
  return error;
}

static int get_commit_id_by_vol_name(const char *name, char **cid) {
  int error = 0;
  struct elton_rpc_ns ns;
  struct get_volume_id_request req;
  struct get_volume_id_response *res;

  RETURN_IF(server.ops->new_session(&server, &ns, NULL));
  req.volume_name = name;
  GOTO_IF(free_ns, ns.ops->send_struct(&ns, GET_VOLUME_ID_REQUEST_ID, &req));
  GOTO_IF(free_ns,
          ns.ops->recv_struct(&ns, GET_VOLUME_ID_RESPONSE_ID, (void **)&res));
  RETURN_IF(ns.ops->close(&ns));

  GOTO_IF(free_res, get_commit_id_by_vid(res->volume_id, cid));
free_res:
  elton_rpc_free_decoded_data(res);
  return 0;

free_ns:
  WARN_IF(ns.ops->close(&ns));
  return error;
}

int get_commit_id_by_config(struct eltonfs_config *config, char **cid) {
  int error = 0;
  ASSERT_NOT_NULL(config);

  if (config->cid) {
    RETURN_IF(dup_string(cid, config->cid));
    return 0;
  } else if (config->vid) {
    RETURN_IF(get_commit_id_by_vid(config->vid, cid));
    return 0;
  } else if (config->vol_name) {
    RETURN_IF(get_commit_id_by_vol_name(config->vol_name, cid));
    return 0;
  } else {
    ERR("invalid config: Need a parameter either cid or vid or vol_name");
    BUG();
  }

  // Unreachable
}
