package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
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
			if v.Tag != "itemizedlist" && v.Tag != "orderedlist" && v.Tag != "variablelist" {
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

		comment := tview.NewTextView().
			SetTextColor(tcell.ColorGray).
			SetRegions(true).
			SetWordWrap(true)
		comment.SetBorder(false)
		fmt.Fprintf(comment, tc[1])

		text := tview.NewTextView().
			SetTextColor(tcell.ColorWhite).
			SetRegions(false).
			SetWordWrap(false)
		text.SetBorder(false)
		fmt.Fprintf(text, tc[0])

		flex := tview.NewFlex().
			AddItem(text, 0, 1, false).
			AddItem(comment, 0, 1, true)
		buf := fmt.Sprintf("page-%d", page+1)
		title := fmt.Sprintf("%s : %s", buf, tc[2])
		comment.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				if pages.HasPage(buf) {
					pages.SwitchToPage(buf)
					header.SetText(title)
				} else {
					app.Stop()
				}
			}
		})
		pages.AddPage(fmt.Sprintf("page-%d", page), flex, true, true)
	}
	pages.SwitchToPage("page-0")
	header.SetText(fmt.Sprintf("page-0 : %s", tcList[0][2]))
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
