//go:build darwin

package utils

import (
    "os"
    "syscall"
)

func ExtractCTime(info os.FileInfo) (int, int) {
    stat := info.Sys().(*syscall.Stat_t)
    return int(stat.Ctimespec.Sec), int(stat.Ctimespec.Nsec % 1e9)
}

type DevIno struct {
    Dev uint64
    Ino uint64
}

func GetDevIno(path string) (DevIno, error) {
    info, err := os.Stat(path)
    if err != nil {
        return DevIno{}, err
    }

    st := info.Sys().(*syscall.Stat_t)
    return DevIno{
        Dev: uint64(st.Dev),
        Ino: uint64(st.Ino),
    }, nil
}

type GidUid struct {
    Gid uint32
    Uid uint32
}

func GetGidUid(path string) GidUid {
    info, err := os.Stat(path)
    if err != nil {
        return GidUid{}
    }

    st := info.Sys().(*syscall.Stat_t)
    return GidUid{
        Gid: st.Gid,
        Uid: st.Uid,
    }
}
