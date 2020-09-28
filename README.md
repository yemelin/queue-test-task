##Queue test task

attempt to implement  
![the design](./docs/scheme.jpg)  
[description](./docs/task.md)


### Run
`go build`  
`./queue_test_task --config <config_file>`  

###Run in Docker
`CONFIG=<abs_path_to_config> make run`  
or  
`CONFIG=<abs_path_to_config> DEST=<abs_path_to_dir> make run`  
choose the latter to map the output directory to `DEST` 
if storage type is 1 in the config. program output will be in
file `data.txt`
stop gracefully with `make stop`
CTRL-C will stop execution gracefully, but return exit code 130