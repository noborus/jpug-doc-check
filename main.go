package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// tcList type
const (
	TEXT = iota
	COMMENT
	PATH
)

func elementToString(elem *etree.Element) (ret string) {
	doc := etree.NewDocument()
	c := elem.Copy()
	doc.AddChild(c)
	ret, err := doc.WriteToString()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WriteToString() Error:%s", err)
		return
	}
	return ret
}

func getTextComment(i int, token []etree.Token) (text string, comment string) {
	for _, c := range token {
		switch v := c.(type) {
		case *etree.Element:
			if v.Tag != "itemizedlist" && v.Tag != "orderedlist" && v.Tag != "variablelist" && v.Tag != "blockquote" {
				text += elementToString(v)
			}
		case *etree.CharData:
			text += v.Data
		case *etree.Comment:
			comment += v.Data
		default:
			fmt.Fprintf(os.Stderr, "Unknown:%T", v)
		}
	}
	return strings.TrimLeft(text, " \n"), strings.TrimLeft(comment, "\n")
}

func xmlParse(root *etree.Element, filename string) [][]string {
	var tcList [][]string
	for i, child := range root.ChildElements() {
		if child.Tag == "para" {
			token := child.Child
			text, comment := getTextComment(i, token)
			path := filename + " : " + child.GetPath()
			tcList = append(tcList, []string{text, comment, path})
		} else {
			tcList = append(tcList, xmlParse(child, filename)...)
		}
	}
	return tcList
}

func draw(tcList [][]string) {
	app := tview.NewApplication()
	grid := tview.NewGrid().
		SetRows(2, 0).
		SetBorders(false)
	pages := tview.NewPages()
	header := tview.NewTextView()
	header.SetTextAlign(tview.AlignCenter)
	grid.AddItem(header, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(pages, 1, 0, 1, 1, 0, 0, true)
	for page, tc := range tcList {

		text := tview.NewTextView().
			SetTextColor(tcell.ColorWhite).
			SetRegions(false).
			SetWordWrap(false)
		text.SetBorder(false)
		fmt.Fprintf(text, tc[TEXT])

		comment := tview.NewTextView().
			SetTextColor(tcell.ColorWhiteSmoke).
			SetRegions(true).
			SetWordWrap(true)
		comment.SetBorder(false)
		fmt.Fprintf(comment, tc[COMMENT])

		flex := tview.NewFlex().
			AddItem(text, 0, 1, true).
			AddItem(comment, 0, 1, false)
		prev := fmt.Sprintf("page-%d", page-1)
		prevTitle := fmt.Sprintf("%s : %s", prev, tc[PATH])
		next := fmt.Sprintf("page-%d", page+1)
		nextTitle := fmt.Sprintf("%s : %s", next, tc[PATH])
		text.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				app.Stop()
			}
			if key == tcell.KeyEnter || key == tcell.KeyTab {
				if pages.HasPage(next) {
					pages.SwitchToPage(next)
					header.SetText(nextTitle)
				} else {
					app.Stop()
				}
			}
			if key == tcell.KeyBacktab {
				if pages.HasPage(prev) {
					pages.SwitchToPage(prev)
					header.SetText(prevTitle)
				}
			}
		})
		pages.AddPage(fmt.Sprintf("page-%d", page), flex, true, true)
	}
	pages.SwitchToPage("page-0")
	header.SetText(fmt.Sprintf("page-0 : %s", tcList[0][PATH]))
	app.SetFocus(pages)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}

func main() {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(os.Args[1]); err != nil {
		panic(err)
	}
	// _ = xmlParse(doc.Root())
	tcList := xmlParse(doc.Root(), os.Args[1])
	draw(tcList)
}
