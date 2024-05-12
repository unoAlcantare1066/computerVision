extern int goWriteCallback(void *, unsigned char *, int);
extern int goCloseCallback(void *);

int CallWriteCb(void* userData, unsigned char *data, int len) {
    return goWriteCallback(userData, data, len);
}

int CallCloseCb(void* userData) {
    return goCloseCallback(userData);
}