package main

import (
	"fmt"
	"os"

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

func xmlParse(root *etree.Element)([][]string) {
	fmt.Printf("%s\n",root.Tag)
	var ct [][]string
	for _, para := range root.FindElements("//*/para") {
		fmt.Printf("%s\n",para)
		ch := para.Child
		d := []string{"", ""}
		for _, c := range ch {
			switch v := c.(type) {
			case *etree.Element:
				d[0] += elementToString(v)
				// TODO: v.Tag != "itemizedlist" && v.Tag != "orderedlist" && v.Tag != "variablelist"{
			case *etree.CharData:
				d[0] += v.Data
			case *etree.Comment:
				d[1] += v.Data
			default:
				fmt.Fprintf(os.Stderr, "Unknown:%T", v)
			}
		}
		ct = append(ct, d)
	}
	return ct
}

func main() {
	app := tview.NewApplication()
	pages := tview.NewPages()
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(os.Args[1]); err != nil {
		panic(err)
	}
	ct := xmlParse(doc.Root())
	for page, h := range ct {
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
