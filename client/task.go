package client

import (
	"container/list"
	"sync"

	"github.com/yangchenxing/go-dds/file"
)

import "container/heap"

const (
	dutyPriorityThreshold = 9
)

type task struct {
	key        string
	blockID    int
	fromServer bool
	file       *file.File
}

type PriorityDecider interface {
	Decide(id int) (priority int, fromServer bool)
}

type priorityBlock struct {
	id         int
	priority   int
	fromServer bool
}

type priorityBlockSlice struct {
	priorityDecider PriorityDecider
	blocks          []priorityBlock
}

func (slice *priorityBlockSlice) Len() int {
	return len(slice.blocks)
}

func (slice *priorityBlockSlice) Less(i, j int) bool {
	return slice.blocks[i].priority < slice.blocks[j].priority
}

func (slice priorityBlockSlice) Swap(i, j int) {
	temp := slice.blocks[i]
	slice.blocks[i] = slice.blocks[j]
	slice.blocks[j] = temp
}

func (slice *priorityBlockSlice) Push(x interface{}) {
	id := x.(int)
	priority, fromServer := slice.priorityDecider.Decide(id)
	item := priorityBlock{
		priority:   priority,
		fromServer: fromServer,
		id:         id,
	}
	slice.blocks = append(slice.blocks, item)
}

func (slice *priorityBlockSlice) Pop() interface{} {
	n := len(slice.blocks)
	item := slice.blocks[n-1]
	slice.blocks = slice.blocks[:n-1]
	return item
}

type taskGenerator struct {
	sync.Mutex
	key    string
	file   *file.File
	blocks *priorityBlockSlice
}

func newTaskGenerator(size int, priorityDecider PriorityDecider) *taskGenerator {
	generator := &taskGenerator{
		blocks: &priorityBlockSlice{
			priorityDecider: priorityDecider,
			blocks:          make([]priorityBlock, 0, size),
		},
	}
	for i := 0; i < size; i++ {
		heap.Push(generator.blocks, i)
	}
	return generator
}

func (generator *taskGenerator) getTask() *task {
	block := heap.Pop(generator.blocks).(priorityBlock)
	return &task{
		blockID:    block.id,
		fromServer: block.fromServer,
		file:       generator.file,
	}
}

type taskGenerators struct {
	sync.Mutex
	outCh      chan *task
	inCh       chan *taskGenerator
	generators *list.List
}

func newTaskGenerators() *taskGenerators {
	generators := &taskGenerators{
		outCh:      make(chan *task),
		inCh:       make(chan *taskGenerator),
		generators: list.New(),
	}
	go generators.run()
	return generators
}

func (generators *taskGenerators) run() {
	for {
		select {
		case generator := <-generators.inCh:
			generators.generators.PushBack(generator)
		default:
			if generators.generators.Len() == 0 {
				generators.generators.PushBack(<-generators.inCh)
			}
			generator := generators.generators.Front().Value.(*taskGenerator)
			task := generator.getTask()
			if task == nil {
				generators.generators.Remove(generators.generators.Front())
			} else {
				generators.outCh <- task
			}
		}
	}
}

func (generators *taskGenerators) Chan() <-chan *task {
	return generators.outCh
}

func (generators *taskGenerators) addGenerator(generator *taskGenerator) {
	generators.inCh <- generator
}
