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

func xmlParse(root *etree.Element) [][]string {
	var tcList [][]string
	for i, child := range root.ChildElements() {
		if child.Tag == "para" {
			token := child.Child
			text, comment := getTextComment(i, token)
			tcList = append(tcList, []string{text, comment})
		} else {
			tcList = append(tcList, xmlParse(child)...)
		}
	}
	return tcList
}

func draw(tcList [][]string) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	for page, h := range tcList {
		comment := tview.NewTextView().
			SetTextColor(tcell.ColorGreen).
			SetRegions(true).
			SetWordWrap(true)
		text := tview.NewTextView().
			SetTextColor(tcell.ColorWhite).
			SetRegions(false).
			SetWordWrap(false)
		comment.SetBorder(false)
		text.SetBorder(false)
		fmt.Fprintf(text, h[0])
		fmt.Fprintf(comment, h[1])
		flex := tview.NewFlex().
			AddItem(text, 0, 1, false).
			AddItem(comment, 0, 1, true)
		buf := fmt.Sprintf("page-%d", page+1)
		comment.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				pages.SwitchToPage(buf)
			}
		})
		pages.AddPage(fmt.Sprintf("page-%d", page), flex, true, true)
	}
	pages.SwitchToPage("page-0")
	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}
func main() {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(os.Args[1]); err != nil {
		panic(err)
	}
	// _ = xmlParse(doc.Root())
	tcList := xmlParse(doc.Root())
	draw(tcList)
}
