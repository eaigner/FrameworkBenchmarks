package controllers

import (
	"benchmark/app/db"
	"github.com/robfig/revel"
	"math/rand"
	"runtime"
	"sort"
	"sync"
)

type MessageStruct struct {
	Message string `json:"message"`
}

type World struct {
	Id           uint16 `json:"id" sql:"pk"`
	RandomNumber uint16 `json:"randomNumber"`
}

type Fortune struct {
	Id      uint16 `json:"id" sql:"pk"`
	Message string `json:"message"`
}

const WorldRowCount = 10000

func init() {
	revel.OnAppStart(func() {
		runtime.GOMAXPROCS(runtime.NumCPU())
		db.Init()
	})
}

type App struct {
	*revel.Controller
}

func (c App) Json() revel.Result {
	c.Response.ContentType = "application/json"
	return c.RenderJson(MessageStruct{"Hello, world"})
}

func (c App) Db(queries int) revel.Result {
	if queries <= 1 {
		hd := db.Hood.Copy()
		w := make([]World, 1)
		err := hd.Where("id", "=", rand.Intn(WorldRowCount)+1).
			Find(&w)
		if err != nil {
			revel.ERROR.Fatalf("Error scanning world row: %v", err)
		}
		return c.RenderJson(w)
	}

	ww := make([]*World, queries)
	var wg sync.WaitGroup
	wg.Add(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			hd := db.Hood.Copy()
			ww[i] = new(World)
			err := hd.Where("id", "=", rand.Intn(WorldRowCount)+1).Find(ww[i])
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return c.RenderJson(ww)
}

func (c App) Update(queries int) revel.Result {
	if queries <= 1 {
		hd := db.Hood.Copy()
		w := make([]World, 1)
		err := hd.Where("id", "=", rand.Intn(WorldRowCount)+1).Find(&w)
		w[0].RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
		_, err = hd.Save(&w[0])
		if err != nil {
			revel.ERROR.Fatalf("Error updating row: %s", err)
		}
		return c.RenderJson(w[0])
	}

	var (
		ww = make([]*World, queries)
		wg sync.WaitGroup
	)
	wg.Add(queries)
	for i := 0; i < queries; i++ {
		go func(i int) {
			hd := db.Hood.Copy()
			ww[i] = new(World)
			err := hd.Where("id", "=", rand.Intn(WorldRowCount)+1).Find(ww[i])
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			ww[i].RandomNumber = uint16(rand.Intn(WorldRowCount) + 1)
			_, err = hd.Save(ww[i])
			if err != nil {
				revel.ERROR.Fatalf("Error scanning world row: %v", err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	return c.RenderJson(ww)
}

func (c App) Fortune() revel.Result {
	hd := db.Hood.Copy()
	var fortunes Fortunes
	err := hd.Find(&fortunes)
	if err != nil {
		revel.ERROR.Fatalf("Error selecting fortunes: %v", err)
	}
	fortunes = append(fortunes,
		&Fortune{Message: "Additional fortune added at request time."})

	sort.Sort(ByMessage{fortunes})
	return c.Render(fortunes)
}

type Fortunes []*Fortune

func (s Fortunes) Len() int      { return len(s) }
func (s Fortunes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByMessage struct{ Fortunes }

func (s ByMessage) Less(i, j int) bool { return s.Fortunes[i].Message < s.Fortunes[j].Message }
