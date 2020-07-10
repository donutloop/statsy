package api

import (
	"database/sql"
	"fmt"
	"github.com/donutloop/statsy/internal/dao"
	"github.com/donutloop/statsy/internal/handler"
	"github.com/donutloop/statsy/internal/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func NewAPI(test bool) *API {
	return &API{
		Test: test,
	}
}

type API struct {
	addrs  string
	Server *server.Server
	Test   bool
}

func (a *API) Bootstrap() {
	err := godotenv.Load(os.Getenv("SERVICE_ENV_FILE"))
	if err != nil {
		log.Fatal("error loading services.env file")
	}

	a.addrs = os.Getenv("SERVER_ADDRS")
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlDB := os.Getenv("MYSQL_DATABASE")
	mysqlHOST := os.Getenv("MYSQL_HOST")

	logrus.Info("connect to mysql")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", mysqlUser, mysqlPassword, mysqlHOST, mysqlDB)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	d := dao.New(db)
	a.Server = server.New(d)

	stubStats := handler.HandlerDomainFunc(handler.HandleCustomer)

	getStatsByCustomerID := handler.HandlerDomainFunc(handler.GetStatsByCustomerID)

	a.Server.AddPostHandlerWithStats("/customer/stats", stubStats)
	a.Server.AddGetHandler("/customer/stats/{customerID}/day/{day}", getStatsByCustomerID)
}

func (a *API) Start() {

	logrus.Info("start server")
	if err := a.Server.Start(os.Getenv("SERVER_ADDRS"), a.Test); err != nil {
		log.Fatal(fmt.Sprintf("error server could not listen on addr %v, err: %v", a.addrs, err))
	}
}

func (a *API) Stop() {
	a.Server.Stop(a.Test)
}
