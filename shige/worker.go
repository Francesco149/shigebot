/*
	Copyright 2015 Franc[e]sco (lolisamurai@tfwno.gf)
	This file is part of Shigebot.
	Shigebot is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	Shigebot is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with Shigebot. If not, see <http://www.gnu.org/licenses/>.
*/

package shige

import "log"

type Worker struct {
	name       string
	jobs       chan func()
	kill, join chan bool
}

func NewWorker(name string, maxJobs int) *Worker {
	return &Worker{
		name,
		make(chan func(), maxJobs),
		make(chan bool, 1), make(chan bool, 1),
	}
}

func (w *Worker) Start() {
	go w.work()
}

func (w *Worker) Join() {
	<-w.join
}

func (w *Worker) Terminate() {
	close(w.jobs)
}

func (w *Worker) Abort() {
	w.kill <- true
	close(w.kill)
}

func (w *Worker) Do(job func()) {
	w.jobs <- job
}

func (w *Worker) Await(job func()) {
	done := make(chan bool, 1)
	w.Do(func() {
		job()
		done <- true
		close(done)
	})
	<-done
}

func (w *Worker) work() {
	log.Println("Worker", w.name, "started.")
	running := true
	var job func()
	for running {
		select {
		case job, running = <-w.jobs:
			job()
		case <-w.kill:
			running = false
			log.Println("Worker", w.name, "received kill signal.")
		}
	}
	w.join <- true
	close(w.join)
	log.Println("Worker", w.name, "terminated")
}
