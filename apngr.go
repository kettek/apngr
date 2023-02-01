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
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/kettek/apng"
)

// Frame represents frames that can be used in JSON
type Frame struct {
	Numerator   *int
	Denominator *int
	// If the first frame, this marks it as being the default image.
	Default *bool
	// May be "none", "background", or "previous"
	Dispose *string
	// May be "over"
	Blend *string
	// Source image file to use
	Image string
}

// Frames are what is expected from a passed in JSON file.
type Frames []Frame

var varDelayNumerator, varDelayDenominator, varOutputFrameIteratorStart int
var varDisposeString, varBlendString, varOutputFramePadding, varOutputDirectory string
var varFirstDefault, varIterateFirstFrame bool
var varMaintainPaletted bool
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
		maintainPalettedDefault         = false
		maintainPalettedUsage           = "convert: maintain paletted and do not convert to RGBA if the palette entry count exceeds 256"
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
	flag.BoolVar(&varMaintainPaletted, "maintainPaletted", varMaintainPaletted, maintainPalettedUsage)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  animate|a [options] [out.png] [frame1.png frames.json ... frameN.png]\n")
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
		Frames:    make([]apng.Frame, 0),
	}

	makeFrame := func() apng.Frame {
		frame := apng.Frame{}
		switch varDisposeString {
		case "none":
			frame.DisposeOp = apng.DISPOSE_OP_NONE
		case "background":
			frame.DisposeOp = apng.DISPOSE_OP_BACKGROUND
		case "previous":
			frame.DisposeOp = apng.DISPOSE_OP_PREVIOUS
		default:
			frame.DisposeOp = apng.DISPOSE_OP_BACKGROUND
		}
		switch varBlendString {
		case "over":
			frame.BlendOp = apng.BLEND_OP_OVER
		}
		frame.DelayDenominator = uint16(varDelayDenominator)
		frame.DelayNumerator = uint16(varDelayNumerator)
		return frame
	}

	firstFrameRead := false
	for i := 2; i < len(args); i = i + 1 {
		f, err := os.Open(args[i])
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if strings.HasSuffix(args[i], ".json") {
			var frames Frames
			d := json.NewDecoder(f)
			if err := d.Decode(&frames); err != nil {
				panic(err)
			}
			fmt.Println("AIGHT", frames)
			for i, f := range frames {
				frame := makeFrame()

				if f.Image == "" {
					panic(fmt.Errorf("unprovided image for %s:%d", args[i], i))
				}
				fmt.Printf("... %s\n", f.Image)

				img, err := os.Open(f.Image)
				if err != nil {
					panic(err)
				}
				defer img.Close()

				m, err := png.Decode(img)
				if err != nil {
					panic(err)
				}
				frame.Image = m

				if f.Dispose != nil {
					switch *f.Dispose {
					case "none":
						frame.DisposeOp = apng.DISPOSE_OP_NONE
					case "background":
						frame.DisposeOp = apng.DISPOSE_OP_BACKGROUND
					case "previous":
						frame.DisposeOp = apng.DISPOSE_OP_PREVIOUS
					default:
						frame.DisposeOp = apng.DISPOSE_OP_BACKGROUND
					}
				}
				if f.Blend != nil {
					if *f.Blend == "over" {
						frame.BlendOp = apng.BLEND_OP_OVER
					}
				}
				if f.Denominator != nil {
					frame.DelayDenominator = uint16(*f.Denominator)
				}
				if f.Numerator != nil {
					frame.DelayNumerator = uint16(*f.Numerator)
				}
				if !firstFrameRead {
					if f.Default != nil {
						frame.IsDefault = *f.Default
					} else {
						frame.IsDefault = varFirstDefault
					}
					firstFrameRead = true
				}
				a.Frames = append(a.Frames, frame)
			}
		} else {
			frame := makeFrame()
			if !firstFrameRead {
				frame.IsDefault = varFirstDefault
				firstFrameRead = true
			}
			m, err := png.Decode(f)
			if err != nil {
				panic(err)
			}
			frame.Image = m
			a.Frames = append(a.Frames, frame)
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

			// Find if we need to repalettize.
			maxPaletteLength := -1
			mustRepalettize := false
			for i := 0; i < len(g.Image); i++ {
				if maxPaletteLength == -1 {
					maxPaletteLength = len(g.Image[i].Palette)
				}
				if len(g.Image[i].Palette) != maxPaletteLength {
					mustRepalettize = true
					break
				}
				for j := 0; j < len(g.Image); j++ {
					if j == i {
						continue
					}
					for k := 0; k < len(g.Image[i].Palette); k++ {
						r1, g1, b1, a1 := g.Image[i].Palette[k].RGBA()
						r2, g2, b2, a2 := g.Image[j].Palette[k].RGBA()
						if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
							mustRepalettize = true
							break
						}
					}
					if mustRepalettize {
						break
					}
				}
				if mustRepalettize {
					break
				}
			}

			// Analyze for RGBA upgrade
			uniqueColors := make(map[[4]uint32]struct{})
			for _, img := range g.Image {
				for _, p := range img.Palette {
					r, g, b, a := p.RGBA()
					k := [4]uint32{r, g, b, a}
					if _, ok := uniqueColors[k]; !ok {
						uniqueColors[k] = struct{}{}
					}
					if len(uniqueColors) > 256 {
						break
					}
				}
				if len(uniqueColors) > 256 {
					break
				}
			}

			var frames []image.Image

			if len(uniqueColors) > 256 && !varMaintainPaletted {
				// Convert frames to APNG
				for _, frame := range g.Image {
					img := image.NewRGBA(frame.Bounds())
					draw.Draw(img, frame.Bounds(), frame, frame.Rect.Min, draw.Src)
					frames = append(frames, img)
				}
			} else if mustRepalettize {
				// Repalettize if we have less than 256 colors.
				insertColorEntry := func(f *image.Paletted, c color.Color) (int, error) {
					r1, g1, b1, a1 := c.RGBA()
					// Use existing entry if it exists.
					for entryIndex, entry := range f.Palette {
						r2, g2, b2, a2 := entry.RGBA()
						if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
							return entryIndex, nil
						}
					}
					// No such color exists, attempt to insert.
					if len(f.Palette) >= 256 {
						return -1, &PaletteLimitError{}
					}
					f.Palette = append(f.Palette, c)
					return len(f.Palette) - 1, nil
				}
				getClosestColorEntry := func(f *image.Paletted, c color.Color) int {
					return f.Palette.Index(c)
				}

				primaryFrame := g.Image[0]
				frames = append(frames, primaryFrame)

				for _, frame := range g.Image {
					if frame == primaryFrame {
						continue
					}
					m := make(map[int]int)
					for i := 0; i < len(frame.Palette); i++ {
						newIndex, err := insertColorEntry(primaryFrame, frame.Palette[i])
						if err == nil {
							m[i] = newIndex
						} else {
							newIndex = getClosestColorEntry(primaryFrame, frame.Palette[i])
							m[i] = newIndex
						}
					}
					// Make new pix referencing frame's pix to remap palette indices.
					p := make([]uint8, len(frame.Pix))
					for x := frame.Bounds().Min.X; x < frame.Bounds().Max.X; x++ {
						for y := frame.Bounds().Min.Y; y < frame.Bounds().Max.Y; y++ {
							i := frame.PixOffset(x, y)
							p[i] = uint8(m[int(frame.Pix[i])])
						}
					}
					frame.Pix = p
					frame.Palette = primaryFrame.Palette
					frames = append(frames, frame)
				}
			} else {
				// Otherwise write out as it is.
				for _, frame := range g.Image {
					frames = append(frames, frame)
				}
			}

			a.LoopCount = uint(g.LoopCount)
			a.Frames = make([]apng.Frame, len(frames))
			fmt.Printf("Total Frames: %d\n", len(frames))
			for i := 0; i < len(frames); i = i + 1 {
				fmt.Printf("Frame %d...", i)
				a.Frames[i].Image = frames[i]
				a.Frames[i].XOffset = g.Image[i].Bounds().Min.X
				a.Frames[i].YOffset = g.Image[i].Bounds().Min.Y
				a.Frames[i].DelayNumerator = uint16(g.Delay[i])
				switch g.Disposal[i] {
				case gif.DisposalNone:
					a.Frames[i].DisposeOp = apng.DISPOSE_OP_NONE
				case gif.DisposalBackground:
					a.Frames[i].DisposeOp = apng.DISPOSE_OP_BACKGROUND
				case gif.DisposalPrevious:
					a.Frames[i].DisposeOp = apng.DISPOSE_OP_PREVIOUS
				}
				a.Frames[i].BlendOp = apng.BLEND_OP_OVER
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

// Errors
type PaletteLimitError struct{}

func (e *PaletteLimitError) Error() string {
	return "palette limit reached"
}
