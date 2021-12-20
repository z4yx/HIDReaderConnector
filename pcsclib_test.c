//
// Copyright © 2021 Yuxiang Zhang
//

#include "pcsclib.h"

#include <stdio.h>
#include <stdlib.h>

int main(int argc, char const *argv[])
{
    ReaderConnectorInit();
    GoInt handle = MakeReaderConnector();
    ReaderOpen(handle);
    ReaderPwrCtl(handle, 1);
    struct ReaderPollA_return pollRet;
    // pollRet = ReaderPollA(handle);
    // if(pollRet.r0 != 0) {
    //     free(pollRet.r3);
    // } else {
    //     free(pollRet.r1);
    //     free(pollRet.r2);
    // }
    pollRet = ReaderPollA(handle);
    if(pollRet.r0 != 0) {
        fprintf(stderr, "Err: %s\n", pollRet.r3);
        free(pollRet.r3);
    } else {
        printf("%s %s\n", pollRet.r1, pollRet.r2);
        free(pollRet.r1);
        free(pollRet.r2);

        ReaderBeep(handle);

        struct ReaderExchangeApduStr_return exchgRet;
        exchgRet = ReaderExchangeApduStr(handle, "00A40000");
        if(exchgRet.r0 != 0) {
            fprintf(stderr, "Err: %s\n", exchgRet.r2);
            free(exchgRet.r2);
        }else{
            printf("%s\n", exchgRet.r1);
            free(exchgRet.r1);
        }
    }
    ReaderClose(handle);
    FreeReaderConnector(handle);
    return 0;
}
