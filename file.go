package main

type File struct {
    Path string
    Hash string `json:"-"`
}
