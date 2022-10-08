package main

import (
	"fmt"
	"sync"
	"time"
)

type Task struct {
	f func() error
}

func NewTask(f func() error) *Task {
	t := Task{
		f: f,
	}
	return &t
}

func (t *Task) Execute() {
	t.f()
}

type Pool struct {
	EntryChan chan *Task
	WorkNum   int
	JobChan   chan *Task
	Wg        *sync.WaitGroup
}

func NewPool(cap int) *Pool {
	p := Pool{
		EntryChan: make(chan *Task),
		JobChan:   make(chan *Task),
		WorkNum:   cap,
		Wg:        &sync.WaitGroup{},
	}
	return &p
}

func (p *Pool) Worker(id int) {
	defer p.Wg.Done()                    //单个goroutine处理完业务，如果测试defer可以跳出for，如开10个协程，设置超时时间，超时或执行完当前任务直接退出for，回收当前go
	timer := time.After(time.Second * 6) // timer
	exitFlag := false
	for {
		select {
		case job, ok := <-p.JobChan:
			if ok {
				job.Execute()
				fmt.Println("work id:", id, "执行成功", job)
				exitFlag = true
			}
		case <-timer:
			fmt.Println("work timeout,id:", id)
			exitFlag = true
		default:
			fmt.Println("work job chan no data,waiting...")
			time.Sleep(1 * time.Second)
		}
		if exitFlag {
			fmt.Println("work done,id:", id)
			break
		}
	}
}

func (p *Pool) Send(t *Task) {
	p.EntryChan <- t
}

func (p *Pool) Run() {
	for i := 0; i < p.WorkNum; i++ {
		p.Wg.Add(1)
		go p.Worker(i)
	}
	timer := time.After(time.Second * 10) // timer,监听输入，一般不做超时，循环接受任务
	exitFlag := false
	for {
		select {
		case task, ok := <-p.EntryChan:
			if ok {
				p.JobChan <- task
				fmt.Println("read task to job:", task)
			}
		case <-timer:
			fmt.Println("timeout no read")
			exitFlag = true
		default:
			fmt.Println("entry chan have no data,in run waiting...")
			time.Sleep(1 * time.Second)
		}
		if exitFlag {
			fmt.Println("break receive task")
			break
		}
	}
	close(p.JobChan)
	close(p.EntryChan)
}

func main() {
	t1 := NewTask(func() error {
		fmt.Println("matrix 11, t1")
		return nil
	})
	t2 := NewTask(func() error {
		fmt.Println("matrix 12,t2")
		time.Sleep(time.Second * 30)
		return nil
	})
	t3 := NewTask(func() error {
		fmt.Println("matrix 13,t3")
		return nil
	})
	t4 := NewTask(func() error {
		fmt.Println("matrix 14,t4")
		time.Sleep(time.Second * 6)
		return nil
	})
	p := NewPool(10)
	go func() {
		p.Send(t1)
		p.Send(t2)
		p.Send(t3)
		p.Send(t4)
	}()
	p.Run()
	p.Wg.Wait()
	fmt.Println("all done...")
}
