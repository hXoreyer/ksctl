package ctl

import "fmt"

type KBar struct {
	percent int64  //百分比
	cur     int64  //当前进度位置
	total   int64  //总进度
	rate    string //进度条
	graph   string //显示符号
	text    string //后缀内容
}

func (bar *KBar) New(start, total int64, text string) {
	bar.cur = start
	bar.total = total
	bar.text = text
	if bar.graph == "" {
		bar.graph = "█"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph //初始化进度条位置
	}
}

func (bar *KBar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		for i := 0; i < int(bar.percent-last)/2; i++ {
			bar.rate += bar.graph
		}
	} else {
		bar.percent = last
	}
	fmt.Printf("\r [%-51s]%3d%%  %8d/%-10d[%-s]", bar.rate, bar.percent, bar.cur, bar.total, bar.text)
}

func (bar *KBar) getPercent() int64 {
	pc := int64(float32(bar.cur) / float32(bar.total) * 100)
	return pc
}

func (bar *KBar) NewWithGraph(start, total int64, graph string, text string) {
	bar.graph = graph
	bar.New(start, total, text)
}

func (bar *KBar) Finish() {
	fmt.Println()
}
