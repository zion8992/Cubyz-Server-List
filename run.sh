MYSQL_ROOT_PASSWORD="H0EeLfLnO,xDEVELOPERSx4c!#%"

if [ -z "$(printenv mysqlpass)" ]; then
    echo "mysqlpass not declared in enviroment!"
    mysqlpass="$MYSQL_ROOT_PASSWORD"
fi

go run ./src/ -dbpass "$mysqlpass" "$@"