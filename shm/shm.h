#ifndef SHM_H
#define SHM_H
#include <string.h>
#include <stdlib.h>
#include <sys/shm.h>
#include <sys/types.h>
#include <sys/ipc.h>

#define IPC_KEY_PROJID 0x42

int sysv_shm_open(int size, int flags, int perm);
void *sysv_shm_attach(int shm_id);
int sysv_shm_detach(void *addr);
int sysv_shm_write(int shm_id, void* input, int len, int offset);
int sysv_shm_read(int shm_id, void* output, int len, int offset);
size_t sysv_shm_get_size(int shm_id);
int sysv_shm_lock(int shm_id);
int sysv_shm_unlock(int shm_id);
int sysv_shm_close(int shm_id);

// SHM_H
#endif
