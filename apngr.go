/*
Copyright 2018-2022 Ketchetwahmeegwun T. Southall / kts of kettek. All Rights reserved.

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
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"path"

	flag "github.com/spf13/pflag"

	"github.com/kettek/apng"
)

var varDelayNumerator, varDelayDenominator, varOutputFrameIteratorStart int
var varDisposeString, varBlendString, varOutputFramePadding, varOutputDirectory string
var varFirstDefault, varIterateFirstFrame bool
var varLoopCount int

func main() {
	defer func() {
		if r := recover(); r != nil {
			err, _ := r.(error)
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		}
	}()

	const (
		numeratorDefault                = 1
		numeratorUsage                  = "animate: delay numerator"
		denominatorDefault              = 10
		denominatorUsage                = "animate: delay denominator"
		disposeDefault                  = "background"
		disposeUsage                    = "animate: dispose operation, may be none, background, or previous"
		blendDefault                    = "source"
		blendUsage                      = "animate: blend operation, may be source or over"
		loopDefault                     = 0
		loopUsage                       = "animate: loop count, 0 means infinite"
		firstFrameDefault               = false
		firstFrameUsage                 = "animate: mark the first frame as being the default image"
		firstFrameIsNumber              = "extract: outputs the first frame as part of the extraction iteration instead of \"default.png\""
		outputFrameIteratorStartDefault = 0
		outputFrameIteratorStartUsage   = "extract: the numerical representation of the starting number to represent frames"
		outputFramePaddingDefault       = ""
		outputFramePaddingUsage         = "extract: amount of zeros to pad the frame number with. Defaults to the nearest place value of the frame total"
		outputDirectoryDefault          = ""
		outputDirectoryUsage            = "extract: output directory to extract to. Defaults to the file's name without an extension"
	)
	flag.IntVarP(&varDelayNumerator, "numerator", "n", numeratorDefault, numeratorUsage)
	flag.IntVarP(&varDelayDenominator, "denominator", "d", denominatorDefault, denominatorUsage)
	flag.IntVarP(&varLoopCount, "loopCount", "l", loopDefault, loopUsage)
	flag.StringVar(&varDisposeString, "dispose", disposeDefault, disposeUsage)
	flag.StringVar(&varBlendString, "blend", blendDefault, blendUsage)
	flag.BoolVar(&varFirstDefault, "firstDefault", firstFrameDefault, firstFrameUsage)
	flag.StringVarP(&varOutputDirectory, "output", "o", outputDirectoryDefault, outputDirectoryUsage)
	flag.IntVarP(&varOutputFrameIteratorStart, "iteratorStart", "i", outputFrameIteratorStartDefault, outputFrameIteratorStartUsage)
	flag.StringVarP(&varOutputFramePadding, "iteratorPadding", "p", outputFramePaddingDefault, outputFramePaddingUsage)
	flag.BoolVar(&varIterateFirstFrame, "iterateDefault", varIterateFirstFrame, firstFrameIsNumber)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  animate|a [options] [out.png] [frame1.png frame2.png ... frameN.png]\n")
		fmt.Fprintf(os.Stderr, "  convert|c [anim1.gif anim2.gif ... animN.gif]\n")
		fmt.Fprintf(os.Stderr, "  extract|e [anim1.png anim2.png ... animN.png]\n")
		fmt.Fprintf(os.Stderr, "  query|q [anim1.png anim2.png ... animN.png]\n")

		fmt.Fprintf(os.Stderr, "\nOptions: \n")
		flag.PrintDefaults()
	}
	flag.CommandLine.SortFlags = false
	flag.Parse()

	args := flag.Args()

	if len(args) <= 1 {
		flag.Usage()
		return
	}
	if args[0] == "e" || args[0] == "extract" {
		extract()
	} else if args[0] == "a" || args[0] == "animate" {
		animate()
	} else if args[0] == "c" || args[0] == "convert" {
		convert()
	} else if args[0] == "q" || args[0] == "query" {
		query()
	}
}

func extract() {
	args := flag.Args()
	for i := 1; i < len(args); i = i + 1 {
		fmt.Printf("Parsing %s\n", args[i])
		f, err := os.Open(args[i])
		if err != nil {
			panic(err)
		}
		defer f.Close()

		a, err := apng.DecodeAll(f)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Extracting %d frames!\n", len(a.Frames))
		var fdir string
		if varOutputDirectory != "" {
			fdir = varOutputDirectory
		} else {
			fdir = path.Base(args[i])
			ext := path.Ext(args[i])
			if len(ext) != 0 {
				fdir = fdir[:len(fdir)-len(ext)]
			}
		}
		fmt.Printf("Creating export directory, \"%s\"\n", fdir)
		err = os.Mkdir(fdir, 0777)
		if err != nil {
			if err == os.ErrExist {
				fmt.Printf("Exists")
			}
			panic(err)
		}

		var outname string
		if varOutputFramePadding != "" {
			outname = fmt.Sprintf("%%0%sd.png", varOutputFramePadding)
		} else {
			outname = fmt.Sprintf("%%0%dd.png", len(fmt.Sprintf("%d", len(a.Frames))))
		}

		for i, frame := range a.Frames {
			fname := ""
			if frame.IsDefault && !varIterateFirstFrame {
				fname = "default.png"
				i--
			} else {
				fname = fmt.Sprintf(outname, varOutputFrameIteratorStart+i)
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
	args := flag.Args()
	if len(args) <= 2 {
		flag.Usage()
	}
	fmt.Println(args)
	outf, err := os.Create(args[1])
	if err != nil {
		panic(err)
	}
	a := apng.APNG{
		LoopCount: uint(varLoopCount),
		Frames:    make([]apng.Frame, len(args)-2),
	}
	for i := 2; i < len(args); i = i + 1 {
		fmt.Printf("Adding %s...\n", args[i])
		f, err := os.Open(args[i])
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := png.Decode(f)
		if err != nil {
			panic(err)
		}
		a.Frames[i-2].Image = m
		switch varDisposeString {
		case "none":
			a.Frames[i-2].DisposeOp = apng.DISPOSE_OP_NONE
		case "background":
			a.Frames[i-2].DisposeOp = apng.DISPOSE_OP_BACKGROUND
		case "previous":
			a.Frames[i-2].DisposeOp = apng.DISPOSE_OP_PREVIOUS
		default:
			a.Frames[i-2].DisposeOp = apng.DISPOSE_OP_BACKGROUND
		}
		switch varBlendString {
		case "over":
			a.Frames[i-2].BlendOp = apng.BLEND_OP_OVER
		}
		a.Frames[i-2].DelayDenominator = uint16(varDelayDenominator)
		a.Frames[i-2].DelayNumerator = uint16(varDelayNumerator)
		if i-2 == 0 {
			a.Frames[i-2].IsDefault = varFirstDefault
		}
	}
	err = apng.Encode(outf, a)
	if err != nil {
		panic(err)
	}
}

func convert() {
	args := flag.Args()
	for i := 1; i < len(args); i = i + 1 {
		fname := path.Base(args[i])
		ext := path.Ext(args[i])
		if len(ext) != 0 {
			fname = fname[:len(fname)-len(ext)]
		}
		fname = fname + ".png"

		fmt.Printf("Attempting to convert %s to %s\n", args[i], fname)

		a := apng.APNG{
			Frames: make([]apng.Frame, 1),
		}

		f, err := os.Open(args[i])
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
	args := flag.Args()
	for i := 1; i < len(args); i = i + 1 {
		fmt.Printf("Parsing %s\n", args[i])
		f, err := os.Open(args[i])
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
			switch frame.Image.(type) {
			case *image.Paletted:
				fmt.Printf("\tImage: Paletted\n")
			case *image.Gray:
				fmt.Printf("\tImage: Gray\n")
			case *image.RGBA:
				fmt.Printf("\tImage: RGBA\n")
			case *image.NRGBA:
				fmt.Printf("\tImage: NRGBA\n")
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
