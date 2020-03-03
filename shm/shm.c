#include "shm.h"

int sysv_shm_open(size_t size, int flags, int perm) {
    int shm_id;

    if(size) {
        // unless otherwise specified, segment is owner-read/write (no exec)
        if(!perm){
            perm = 0600;
        }

        return shmget(IPC_PRIVATE, size, flags|perm);
    } else {
        return shmget(IPC_PRIVATE, size, 0);
    }
}

int sysv_shm_write(int shm_id, void* input, int len, int offset) {
    // attach to the given segment to get its memory address
    char* addr = sysv_shm_attach(shm_id);

    if(addr == (char*)(-1)){
        return -1;
    }

    // copy len bytes from input into addr
    memcpy(addr+offset, input, len);

    // detach
    sysv_shm_detach(addr);

    return 0;
}

void *sysv_shm_attach(int shm_id) {
    return shmat(shm_id, NULL, 0);
}

int sysv_shm_detach(void *addr) {
    return shmdt(addr);
}

int sysv_shm_read(int shm_id, void* output, int len, int offset) {
    // attach to the given segment to get its memory address
    char* addr = sysv_shm_attach(shm_id);

    if(addr == (char*)(-1)){
        return -1;
    }

    // copy len bytes from addr into output
    memcpy(output, addr+offset, len);

    // detach
    sysv_shm_detach(addr);

    return 0;
}

int sysv_shm_lock(int shm_id) {
    return shmctl(shm_id, SHM_LOCK, NULL);
}

int sysv_shm_unlock(int shm_id) {
    return shmctl(shm_id, SHM_UNLOCK, NULL);
}

int sysv_shm_close(int shm_id) {
    return shmctl(shm_id, IPC_RMID, NULL);
}

size_t sysv_shm_get_size(int shm_id) {
    struct shmid_ds shm;

    if(shmctl(shm_id, IPC_STAT, &shm) >= 0) {
        return shm.shm_segsz;
    }else{
        return -1;
    }
}
