/*
Copyright 2018 Ketchetwahmeegwun T. Southall / kts of kettek. All Rights reserved.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
  "github.com/kettek/apng"
  "image"
  "image/png"
  "image/gif"
  "os"
  "path"
  "fmt"
)

func usage() {
  fmt.Printf("Usage:\n")
  fmt.Printf("\t%s [query|q] anim1.png anim2.png ...\n", os.Args[0])
  fmt.Printf("\t%s [convert|c] anim.gif ...\n", os.Args[0])
  fmt.Printf("\t%s [extract|e] frame1.png frame2.png ...\n", os.Args[0])
  fmt.Printf("\t%s [animate|a] output.png frame1.png frame2.png...\n", os.Args[0])
  os.Exit(0)
}

func main() {
  defer func() {
    if r := recover(); r != nil {
      err, _ := r.(error)
      fmt.Fprintf(os.Stderr, "%s\n", err.Error())
    }
  }()
  if len(os.Args) <= 2 {
    usage()
  }
  if os.Args[1] == "e" || os.Args[1] == "extract" {
    extract()
  } else if os.Args[1] == "a" || os.Args[1] == "animate" {
    animate()
  } else if os.Args[1] == "c" || os.Args[1] == "convert" {
    convert()
  } else if os.Args[1] == "q" || os.Args[1] == "query" {
    query()
  }
}

func extract() {
  for i := 2; i < len(os.Args); i = i +1 {
    fmt.Printf("Parsing %s\n", os.Args[i])
    f, err := os.Open(os.Args[i])
    if err != nil {
      panic(err)
    }
    defer f.Close()

    a, err := apng.DecodeAll(f)
    if err != nil {
      panic(err)
    }

    fmt.Printf("Extracting %d frames!\n", len(a.Frames))
    fdir := path.Base(os.Args[i])
    ext := path.Ext(os.Args[i])
    if len(ext) != 0 {
      fdir = fdir[:len(fdir)-len(ext)]
    }
    fmt.Printf("Creating export directory, \"%s\"\n", fdir)
    err = os.Mkdir(fdir, 0777)
    if err != nil {
      if err == os.ErrExist {
        fmt.Printf("Existsts")
      }
      panic(err)
    }
    for i, frame := range a.Frames {
      fname := ""
      if i == 0 && frame.IsDefault {
        fname = "default.png"
        i = i - 1
      } else {
        fname = fmt.Sprintf("%d.png", i)
      }
      fmt.Printf("%s...", path.Join(fdir, fname))
      file, err := os.Create(path.Join(fdir, fname))
      if err != nil {
        panic(err)
      }
      defer file.Close()
      err = png.Encode(file, frame.Image)
      if err != nil {
        panic(err)
      }
      fmt.Printf("...ok!\n")
    }
  }
}

func animate() {
  if len(os.Args) <= 3 {
    usage()
  }
  outf, err := os.Create(os.Args[2])
  if err != nil {
    panic(err)
  }
  a := apng.APNG{
    LoopCount: 0,
    Frames: make([]apng.Frame, len(os.Args)-3),
  }
  for i := 3; i < len(os.Args); i = i + 1 {
    fmt.Printf("Adding %s...\n", os.Args[i])
    f, err := os.Open(os.Args[i])
    if err != nil {
      panic(err)
    }
    defer f.Close()
    m, err := png.Decode(f)
    if err != nil {
      panic(err)
    }
    a.Frames[i-3].Image = m
    a.Frames[i-3].DisposeOp = apng.DISPOSE_OP_BACKGROUND
  }
  err = apng.Encode(outf, a)
}

func convert() {
  for i := 2; i < len(os.Args); i = i +1 {
    fname := path.Base(os.Args[i])
    ext := path.Ext(os.Args[i])
    if len(ext) != 0 {
      fname = fname[:len(fname)-len(ext)]
    }
    fname = fname+".png"

    fmt.Printf("Attempting to convert %s to %s\n", os.Args[i], fname)

    a := apng.APNG{
      Frames: make([]apng.Frame, 1),
    }

    f, err := os.Open(os.Args[i])
    if err != nil {
      panic(err)
    }
    defer f.Close()

    switch ext {
    case ".gif":
      fmt.Printf("Handling GIF\n")
      g, err := gif.DecodeAll(f)
      if err != nil {
        panic(err)
      }
      a.LoopCount = uint(g.LoopCount)
      a.Frames = make([]apng.Frame, len(g.Image))
      fmt.Printf("Total Frames: %d\n", len(g.Image))
      for i := 0; i < len(g.Image); i = i + 1 {
        fmt.Printf("Frame %d...", i)
        a.Frames[i].Image = image.Image(g.Image[i])
        a.Frames[i].DelayNumerator = uint16(g.Delay[i])
        switch g.Disposal[i] {
        case gif.DisposalNone:
          a.Frames[i].DisposeOp = apng.DISPOSE_OP_NONE
        case gif.DisposalBackground:
          a.Frames[i].DisposeOp = apng.DISPOSE_OP_BACKGROUND
        case gif.DisposalPrevious:
          a.Frames[i].DisposeOp = apng.DISPOSE_OP_PREVIOUS
        }
      fmt.Printf("...ok!\n")
      }
      fmt.Printf("Done!\n")
    default:
      fmt.Printf("Unhandled extension, attempting to use image.Decode...\n")
      m, format, err := image.Decode(f)
      if err != nil {
        panic(err)
      }
      fmt.Printf("Using %s!\n", format)
      a.Frames[0].Image = m
    }
    outf, err := os.Create(fname)
    if err != nil {
      panic(err)
    }
    defer outf.Close()
    err = apng.Encode(outf, a)
    if err != nil {
      panic(err)
    }
  }
}

func query() {
  for i := 2; i < len(os.Args); i = i +1 {
    fmt.Printf("Parsing %s\n", os.Args[i])
    f, err := os.Open(os.Args[i])
    if err != nil {
      panic(err)
    }
    defer f.Close()

    a, err := apng.DecodeAll(f)
    if err != nil {
      panic(err)
    }

    fmt.Printf("Found %d frames!\n", len(a.Frames))
    for i, frame := range a.Frames {
      b := frame.Image.Bounds()
      if i == 0 && frame.IsDefault {
        fmt.Printf("Default Image (not included in animation)\n")
        i = i - 1
      } else {
        fmt.Printf("Frame %d\n", i)
      }
      fmt.Printf("\tWidth x Height: %dx%d\n", b.Max.X, b.Max.Y)
      fmt.Printf("\tXOffset x YOffset: %dx%d\n", frame.XOffset, frame.YOffset)
      fmt.Printf("\tDelay: %f\n", frame.GetDelay())
      switch frame.DisposeOp {
      case apng.DISPOSE_OP_NONE:
        fmt.Printf("\tDispose: None (%d)\n", frame.DisposeOp)
      case apng.DISPOSE_OP_BACKGROUND:
        fmt.Printf("\tDispose: Background (%d)\n", frame.DisposeOp)
      case apng.DISPOSE_OP_PREVIOUS:
        fmt.Printf("\tDispose: Previous (%d)\n", frame.DisposeOp)
      default:
        fmt.Printf("\tDispose: INVALID (%d)\n", frame.DisposeOp)
      }
      switch frame.BlendOp {
      case apng.BLEND_OP_SOURCE:
        fmt.Printf("\tBlend: Source (%d)\n", frame.BlendOp)
      case apng.BLEND_OP_OVER:
        fmt.Printf("\tBlend: Over (%d)\n", frame.BlendOp)
      default:
        fmt.Printf("\tBlend: INVALID (%d)\n", frame.BlendOp)
      }
    }
  }
}
