library(plumber)
r <- plumb("plumber.r")
r$run(host="0.0.0.0", port=strtoi(Sys.getenv("PORT")))
