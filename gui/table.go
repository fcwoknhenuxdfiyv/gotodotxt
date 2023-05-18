package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"gotodotxt/tdt"
)

func main() {
	go func() {
		w := app.NewWindow()
		if err := Run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type (
	C = layout.Context
	D = layout.Dimensions
)

type Row struct {
	Description string
	Priority    string
	Due         time.Time
	Threshold   time.Time
	Recurrence  string
}

type model struct {
	file         *tdt.TaskFile
	th           *material.Theme
	timingWindow time.Duration
	rows         []Row
	frameCounter int
	timingStart  time.Time
	FPS          float64
	ops          op.Ops
	grid         component.GridState
	gtx          layout.Context
}

func Run(w *app.Window) error {
	m := Init("~/Nextcloud/Tasks/eep.txt")
	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			m.Update(e)
			m.View()
			e.Frame(m.gtx.Ops)
			m.frameCounter++
		}
	}
}

func Init(fn string) model {
	m := model{}
	m.th = material.NewTheme(gofont.Collection())
	m.timingWindow = time.Second
	m.rows = []Row{}
	m.frameCounter = 0
	m.timingStart = time.Time{}
	opts := tdt.Opts{
		SortOrder:  "done,priority,due-,threshold-",
		ShowFuture: false,
	}
	m.file = tdt.Read(fn, opts).Sort().Filter()
	for _, t := range m.file.Tasks {
		if t.FilteredOut {
			continue
		}
		m.rows = append(m.rows, Row{
			Description: t.Description,
			Priority:    t.Priority,
		})
		if t.HasDue || t.HasThreshold || t.Recurrence.Period != "" {
			m.rows = append(m.rows, Row{
				Due:        t.Due,
				Threshold:  t.Threshold,
				Recurrence: t.Recurrence.String,
			})
		}
	}
	return m
}

func (m *model) Update(e system.FrameEvent) error {
	m.gtx = layout.NewContext(&m.ops, e)
	op.InvalidateOp{}.Add(m.gtx.Ops)
	if m.timingStart == (time.Time{}) {
		m.timingStart = m.gtx.Now
	}
	if interval := m.gtx.Now.Sub(m.timingStart); interval >= m.timingWindow {
		m.FPS = float64(m.frameCounter) / interval.Seconds()
		fmt.Println(m.FPS)
		m.frameCounter = 0
		m.timingStart = m.gtx.Now
	}
	return nil
}

var headingText = []string{"Title"}

func (m *model) View() D {
	// Configure width based on available space and a minimum size.
	minSize := m.gtx.Dp(unit.Dp(200))
	// border := widget.Border{
	// 	Color: color.NRGBA{A: 255},
	// 	Width: unit.Dp(1),
	// }

	inset := layout.UniformInset(unit.Dp(2))

	// Configure a label styled to be a data element.
	dataLabel := material.Body1(m.th, "")
	dataLabel.Font.Variant = "Mono"
	dataLabel.MaxLines = 1
	dataLabel.Alignment = text.Start

	// Measure the height of a heading row.
	orig := m.gtx.Constraints
	m.gtx.Constraints.Min = image.Point{}
	macro := op.Record(m.gtx.Ops)
	dims := inset.Layout(m.gtx, dataLabel.Layout)
	_ = macro.Stop()
	m.gtx.Constraints = orig

	return component.Grid(m.th, &m.grid).Layout(m.gtx, len(m.rows), 2,
		func(axis layout.Axis, index, constraint int) int {
			widthUnit := max(int(float32(constraint)-100), minSize)
			switch axis {
			case layout.Horizontal:
				switch index {
				case 0:
					return 100
				case 1:
					// return 1500
					// fmt.Println(int(widthUnit))
					return int(widthUnit)
				default:
					return 0
				}
			default:
				return dims.Size.Y
			}
		},
		func(gtx C, row, col int) D {
			return inset.Layout(gtx, func(gtx C) D {
				timing := m.rows[row]
				switch col {
				case 0:
					dataLabel.Text = timing.Priority
				case 1:
					txt := timing.Description
					if txt == "" {
						if !timing.Due.IsZero() {
							txt += "due:" + tdt.YMD(timing.Due) + " "
						}
						if !timing.Threshold.IsZero() {
							txt += "t:" + tdt.YMD(timing.Threshold) + " "
						}
						if timing.Recurrence != "" {
							txt += "rec:" + timing.Recurrence
						}
					}
					dataLabel.Text = txt
				}
				return dataLabel.Layout(gtx)
			})
		},
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
