package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gbbackend1/reguser2/api/handler"
	"github.com/gbbackend1/reguser2/api/routeroapi"
	"github.com/gbbackend1/reguser2/api/server"
	"github.com/gbbackend1/reguser2/app/repos/user"
	"github.com/gbbackend1/reguser2/app/starter"
	"github.com/gbbackend1/reguser2/db/mem/usermemstore"
	"github.com/gbbackend1/reguser2/db/sql/pgstore"
)

func main() {
	if tz := os.Getenv("TZ"); tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			log.Printf("error loading location '%s': %v\n", tz, err)
		}
	}

	// output current time zone
	tnow := time.Now()
	tz, _ := tnow.Zone()
	log.Printf("Local time zone %s. Service started at %s", tz,
		tnow.Format("2006-01-02T15:04:05.000 MST"))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	var ust user.UserStore
	stu := os.Getenv("REGUSER_STORE")

	switch stu {
	case "mem":
		ust = usermemstore.NewUsers()
	case "pg":
		dsn := os.Getenv("PG_DSN")
		pgst, err := pgstore.NewUsers(dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer pgst.Close()
		ust = pgst
	default:
		log.Fatal("unknown REGUSER_STORE = ", stu)
	}

	a := starter.NewApp(ust)
	us := user.NewUsers(ust)
	h := handler.NewHandlers(us)

	rh := routeroapi.NewRouterOpenAPI(h)

	srv := server.NewServer(":"+os.Getenv("PORT"), rh)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go a.Serve(ctx, wg, srv)

	<-ctx.Done()
	cancel()
	wg.Wait()
}
