// +build windows

package utils

import (
    "os"
    "syscall"
)

func ExtractCTime(info os.FileInfo) (int, int) {
    stat := info.Sys().(*syscall.Win32FileAttributeData)
    ns := stat.CreationTime.Nanoseconds()
    return int(ns / 1e9), int(ns % 1e9)
}

type DevIno struct {
    Dev uint64
    Ino uint64
}

func GetDevIno(path string) (DevIno, error) {
    handle, err := syscall.CreateFile(
        syscall.StringToUTF16Ptr(path),
        syscall.GENERIC_READ,
        syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
        nil,
        syscall.OPEN_EXISTING,
        syscall.FILE_ATTRIBUTE_NORMAL,
        0,
    )
    if err != nil {
        return DevIno{}, err
    }
    defer syscall.CloseHandle(handle)

    var data syscall.ByHandleFileInformation
    if err := syscall.GetFileInformationByHandle(handle, &data); err != nil {
        return DevIno{}, err
    }

    dev := uint64(data.VolumeSerialNumber)
    ino := (uint64(data.FileIndexHigh) << 32) | uint64(data.FileIndexLow)

    return DevIno{Dev: dev, Ino: ino}, nil
}

type GidUid struct {
    Gid uint32
    Uid uint32
}

func GetGidUid(path string) GidUid {
    return GidUid{0, 0}
}
