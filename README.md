# __Alessor Service__

# Description
My capstone project for my Bachelors Degree. This is part of a Property Management app. It's main functionality is maintence work scheduling, automation of finances, and payment integration.

# Quickstart 
## Setup server
- if you my teacher I will give you access to the .env and config.yml files
- add the .env file to the root of the project
- add the config.yml file to the config dir in the root directory
- run `make tidy` to make sure all dependencies are installed

### **Now your ready to run the server**

### Run the server binary
- run 'make run-bin' in the root dir
- this builds the program binary then exectues it

### Run the server in development mode
- run `make run-live` in the root dir
- this builds the programs binary then executes it with air for live reloading
- the config sets the server to run on port :8087

### Run it with docker
- **but** I have to finish setting that up so..

### Run tests
-- run `make test` or `make test-cover` in root dir
-- `make test` runs all tests in race condition mode
-- `make test-cover` runs all the test but also displays test coverage

