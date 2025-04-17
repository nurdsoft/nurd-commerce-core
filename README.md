# nurd-commerce-core

## Setting up the environment

### Prerequisites

- Install [golang](https://go.dev/doc/install)
- Install [docker](https://docs.docker.com/engine/install/)

### Project setup

- Checkout the project locally

  ```bash
  git clone https://github.com/nurdsoft/nurd-commerce-core
  ```

- Change the directory to the project folder

  ```bash
  cd nurd-commerce-core
  ```

- To install the necessary libraries & tools run the below command

  ```bash
  make setup
  # this will take some time. Grab a coffee ☕️
  ```

## Application Configuration

The application configuration is in the file `./config.yaml`. Please change the values appropriately based on the environment.

## Environment Variables

You can override the configuration values by setting them as environment values. Below are the environment variables that can override the `config.yaml` values.

Copy these values in a `.env` file.

```bash
cp .env.sample .env
```

> [!TIP]
> Ask a team member for the values of the environment variables.

After adding the correct values, use `make env` to export the environment variables (only for MacOS and Linux).

## Running the application

## Local Env

### Start Environment

- To start the services that are required for the application (e.g. database) run the below command

  ```bash
  make start-env
  ```

### Build application

- Build the application using the below command

  ```bash
  make nurd-commerce
  ```

### Running application

- Run the application locally using the below command

  ```bash
  make run-dev
  ```

  Access the application on `http://localhost:8080` and you should see the `OK` JSON response from the application.
  Access the Swagger API docs on `http://127.0.0.1:8080/docs/swagger/`


### Running the application in the docker container

- Run the application in the docker container using the below command

  ```bash
  make start-app
  ```

### Stop Environment

- To stop all the services run the below command

  ```bash
  make stop-env
  ```

## Kick-start running the whole application

- To run all the services including the application run the below commands

  ```bash
  git clone https://github.com/nurdsoft/nurd-commerce-core
  cd nurd-commerce-core
  make start-all
  ```

## Database migrations

We use [sql-migrate](https://github.com/rubenv/sql-migrate) for database migrations

- To create a new migration

  ```bash
  make name=migration_script_file_name new-migration
  ```

- To apply outstanding migrations

  ```bash
  make migrate
  ```

- To roll back the last migration

  ```bash
  make migrate direction=down
  ```


## Generating Mocs for Unit Testing

We use `mockgen`` to generate mocks for unit testing. To generate mocks, run the below command

```bash
mockgen -source=internal/... -destination=internal/mocks/mock.go -package=package_name
```

