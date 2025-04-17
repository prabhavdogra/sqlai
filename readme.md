### Setup
- Install Go
- Run `go mod tidy`
- Make sure your Docker has atleast 5.5 GB of allocated memory
- Run the docker containers using `docker-compose up -d`

### Check Everything is working as expected
- Check the Ollama is working as expected using the curl `curl http://localhost:11434/`
- Check the database is working and the data is present by running the following commands in the container
    - `psql -U postgres -d postgres`
    - `SELECT * FROM students;`
