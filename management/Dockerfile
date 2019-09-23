FROM heroku/heroku:18
COPY ./bin/server /
COPY ./database/1_data_load.up.sql /
EXPOSE 4000
CMD ["/server"]

