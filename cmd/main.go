package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Z3DRP/lessor-service/config"
	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/factories"
	"github.com/Z3DRP/lessor-service/internal/routes"
	"github.com/Z3DRP/lessor-service/internal/services/alssr"
	"github.com/Z3DRP/lessor-service/internal/services/prfl"
	"github.com/Z3DRP/lessor-service/internal/services/property"
	rentalproperty "github.com/Z3DRP/lessor-service/internal/services/rentalProperty"
	"github.com/Z3DRP/lessor-service/internal/services/task"
	"github.com/Z3DRP/lessor-service/internal/services/usr"
	"github.com/Z3DRP/lessor-service/internal/services/worker"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func run() error {
	configPath := os.Getenv("configPath")
	log.Println("starting server setup...")

	if configPath == "" {
		configPath = "./config"
	}

	var apiConfig, configErr = config.ReadConfig(configPath)

	if configErr != nil {
		crane.DefaultLogger.MustDebug(fmt.Sprintf("an error occurred while reading server config, %v", configErr))
		return configErr
	}

	log.Printf("initializing database...")
	dbConnection, err := dac.DbCon(&apiConfig.DatabaseStore)

	if err != nil {
		crane.DefaultLogger.LogFields(logrus.Fields{
			"msg": "database setup error",
			"err": err,
		})
	}

	dbStore := dac.NewBuilder().SetDB(dbConnection).SetBunDB().Build()

	//dbStore := dac.InitStore(dbConnection)
	// creating alessor service will never return an err so ignore it
	alsrService, _ := factories.ServiceFactory("Alessor", dbStore, crane.DefaultLogger)
	alsrHandler, err := factories.HandlerFactory(alsrService.ServiceName(), alsrService)
	if err != nil {
		return err
	}

	log.Printf("initializing services...")
	// creating usr service will never return err so ignore it
	usrService, _ := factories.ServiceFactory("User", dbStore, crane.DefaultLogger)
	usrHandler, err := factories.HandlerFactory(usrService.ServiceName(), usrService)
	if err != nil {
		return err
	}

	PropertyService, err := factories.ServiceFactory("Property", dbStore, crane.DefaultLogger)
	if err != nil {
		return factories.ErrFailedServiceStart{ServiceName: PropertyService.ServiceName(), Err: err}
	}

	propertyHandler, err := factories.HandlerFactory(PropertyService.ServiceName(), PropertyService)
	if err != nil {
		return fmt.Errorf("failed to create property handler %v", err)
	}

	taskService, err := factories.ServiceFactory("Task", dbStore, crane.DefaultLogger)
	if err != nil {
		return factories.ErrFailedServiceStart{ServiceName: taskService.ServiceName(), Err: err}
	}

	taskHandler, err := factories.HandlerFactory(taskService.ServiceName(), taskService)
	if err != nil {
		log.Printf("task err: %v", err)
		return cmerr.ErrUnexpectedData{Wanted: taskHandler, Got: taskHandler}
	}

	rentalPropertyService, _ := factories.ServiceFactory("Rental Property", dbStore, crane.DefaultLogger)
	rentalPropertyHandler, err := factories.HandlerFactory(rentalPropertyService.ServiceName(), rentalPropertyService)
	if err != nil {
		return factories.ErrFailedServiceStart{ServiceName: rentalPropertyService.ServiceName(), Err: err}
	}

	workerService, _ := factories.ServiceFactory("Worker", dbStore, crane.DefaultLogger)
	workerHandler, err := factories.HandlerFactory(workerService.ServiceName(), workerService)
	if err != nil {
		log.Printf("worker hndlr er %v", err)
		return factories.ErrFailedServiceStart{ServiceName: workerService.ServiceName(), Err: err}
	}

	aHandler, ok := alsrHandler.(alssr.AlessorHandler)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: alssr.AlessorHandler{}, Got: alsrHandler}
	}

	uHandler, ok := usrHandler.(usr.UserHandler)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: prfl.ProfileHandler{}, Got: usrHandler}
	}

	pHandler, ok := propertyHandler.(property.PropertyHandler)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: property.PropertyHandler{}, Got: propertyHandler}
	}

	tHandler, ok := taskHandler.(task.TaskHandler)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: task.TaskHandler{}, Got: taskHandler}
	}

	rpHandler, ok := rentalPropertyHandler.(rentalproperty.RentalPropertyHandler)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: rentalproperty.RentalPropertyHandler{}, Got: rpHandler}
	}

	wHandler, ok := workerHandler.(worker.WorkerHandler)
	if !ok {
		return cmerr.ErrUnexpectedData{Wanted: worker.WorkerHandler{}, Got: wHandler}
	}

	zserver, err := routes.NewServer(&apiConfig.ZServer, aHandler, uHandler, pHandler, tHandler, rpHandler, wHandler)
	if err != nil {
		crane.DefaultLogger.MustDebug(fmt.Sprintf("fatal error creating server, %v", err))
		return err
	}

	log.Println("finished...")
	crane.DefaultLogger.MustDebug("server is live and running")
	log.Println("Server is live and running on 8087")

	if err := zserver.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		crane.DefaultLogger.MustDebug(fmt.Sprintf("fatal server error: %s", err))
		return err
	}

	routes.HandleShutdown(zserver)
	return nil
}

func main() {
	// load .env file this is not the same as structure config loaded in run
	if err := godotenv.Load(); err != nil {
		log.Printf("WARNING no .env file found this may cause problems")
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
