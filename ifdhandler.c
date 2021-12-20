//
// Copyright © 2021 Yuxiang Zhang
//

#include "pcsclib.h"

#include <ifdhandler.h>
#include <reader.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>

const static UCHAR ATR[] = {0x3B, 0xF7, 0x11, 0x00, 0x00, 0x81, 0x31, 0xFE, 0x65,
                            0x43, 0x61, 0x6E, 0x6F, 0x6B, 0x65, 0x79, 0x99};
static int reader_init = 0;
static GoInt reader_handle;

enum {
    ICCState_None,
    ICCState_Polled,
} icc_state;

RESPONSECODE IFDHCreateChannel ( DWORD Lun, DWORD Channel )
{
    printf("IFDHCreateChannel %ld %ld\n", Lun, Channel);
    if(!reader_init) {
        ReaderConnectorInit();
        reader_init = 1;
    }
    reader_handle = MakeReaderConnector();
    ReaderOpen(reader_handle);
    ReaderPwrCtl(reader_handle, 1);
    return IFD_SUCCESS;
}

RESPONSECODE IFDHCloseChannel ( DWORD Lun )
{
    printf("IFDHCloseChannel %ld\n", Lun);
    ReaderClose(reader_handle);
    return IFD_SUCCESS;
}

static RESPONSECODE card_state_change(DWORD Lun, int timeout)
{
    struct timespec spec = {.tv_sec = timeout/1000, .tv_nsec = timeout % 1000 * 1000000ll};
    nanosleep(&spec, NULL);
    return IFD_RESPONSE_TIMEOUT;
}

RESPONSECODE IFDHGetCapabilities ( DWORD Lun, DWORD Tag,
                                   PDWORD Length, PUCHAR Value )
{
    printf("IFDHGetCapabilities %ld %#lx\n", Lun, Tag);
    switch (Tag) {
    case TAG_IFD_ATR:
    case SCARD_ATTR_ATR_STRING:
        *Length = sizeof(ATR);
        memcpy(Value, ATR, *Length);
        break;
    case TAG_IFD_SIMULTANEOUS_ACCESS:
        *Length = 1;
        Value[0] = 1;
        break;
    case TAG_IFD_SLOTS_NUMBER:
        *Length = 1;
        Value[0] = 1;
        break;
    case TAG_IFD_POLLING_THREAD_KILLABLE:
        *Length = 1;
        Value[0] = 1;
        break;
    case TAG_IFD_POLLING_THREAD_WITH_TIMEOUT:
        *Length = sizeof(void*);
        *(void**)Value = (void*)card_state_change;
        break;

    default:
        return IFD_ERROR_TAG;
        break;
    }
    return IFD_SUCCESS;
}

RESPONSECODE IFDHSetCapabilities ( DWORD Lun, DWORD Tag,
                                   DWORD Length, PUCHAR Value )
{

    printf("IFDHSetCapabilities %ld %#lx %ld\n", Lun, Tag, Length);
    return IFD_ERROR_TAG;
}

RESPONSECODE IFDHSetProtocolParameters ( DWORD Lun, DWORD Protocol,
        UCHAR Flags, UCHAR PTS1,
        UCHAR PTS2, UCHAR PTS3)
{

    printf("IFDHSetProtocolParameters %ld %ld %#x\n", Lun, Protocol, Flags);
    if(Protocol != SCARD_PROTOCOL_T1)
        return IFD_PROTOCOL_NOT_SUPPORTED;
    return IFD_SUCCESS;
}

RESPONSECODE IFDHPowerICC ( DWORD Lun, DWORD Action,
                            PUCHAR Atr, PDWORD AtrLength )
{
    printf("IFDHPowerICC %ld Action=%#lx\n", Lun, Action);
    if(Action == IFD_POWER_UP || Action == IFD_RESET) {
        *AtrLength = sizeof(ATR);
        memcpy(Atr, ATR, *AtrLength);
    } else if(Action == IFD_POWER_DOWN) {
    } else {
        return IFD_NOT_SUPPORTED;
    }
    return IFD_SUCCESS;
}

void PollCardAutomatically(void)
{
    if(icc_state==ICCState_None) {
        struct ReaderPollA_return pollRet;
        pollRet = ReaderPollA(reader_handle);
        if(pollRet.r0 != 0) {
            fprintf(stderr, "ReaderPollA Err: %s\n", pollRet.r3);
            free(pollRet.r3);
        } else {
            printf("%s %s\n", pollRet.r1, pollRet.r2);
            free(pollRet.r1);
            free(pollRet.r2);

            ReaderBeep(reader_handle);
            icc_state = ICCState_Polled;
        }
    }

}

RESPONSECODE IFDHTransmitToICC ( DWORD Lun, SCARD_IO_HEADER SendPci,
                                 PUCHAR TxBuffer, DWORD TxLength,
                                 PUCHAR RxBuffer, PDWORD RxLength,
                                 PSCARD_IO_HEADER RecvPci )
{
    RESPONSECODE ret;
    printf("IFDHTransmitToICC %ld T=%ld\n", Lun, SendPci.Protocol);
    RecvPci->Protocol = SendPci.Protocol;
    //SCARD_IO_HEADER::Length is not used according to document

    PollCardAutomatically();

    struct ReaderExchangeApdu_return exchgRet;
    exchgRet = ReaderExchangeApdu(reader_handle, TxBuffer, TxLength);
    if(exchgRet.r0 != 0) {
        fprintf(stderr, "ReaderExchangeApdu Err: %s\n", exchgRet.r3);
        free(exchgRet.r3);
        icc_state = ICCState_None;
        *RxLength = 0;
        ret = IFD_COMMUNICATION_ERROR;
    }else{
        if(exchgRet.r1 > *RxLength) {
            *RxLength = 0;
            ret = IFD_ERROR_INSUFFICIENT_BUFFER;
        }
        else {
            *RxLength = exchgRet.r1;
            memcpy(RxBuffer, exchgRet.r2, exchgRet.r1);
            ret = IFD_SUCCESS;
        }
        free(exchgRet.r2);
    }
    return ret;
}

RESPONSECODE IFDHControl (DWORD Lun, DWORD dwControlCode, PUCHAR
                          TxBuffer, DWORD TxLength, PUCHAR RxBuffer, DWORD RxLength,
                          LPDWORD pdwBytesReturned)
{

    *pdwBytesReturned = 0;
    return IFD_ERROR_NOT_SUPPORTED;
}

RESPONSECODE IFDHICCPresence( DWORD Lun )
{
    printf("IFDHICCPresence %ld\n", Lun);
    PollCardAutomatically();
    return icc_state == ICCState_Polled ? IFD_ICC_PRESENT : IFD_ICC_NOT_PRESENT;
}
