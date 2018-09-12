#' Echo the parameter that was sent in
#' @param msg The message to echo back.
#' @get /
function(msg=""){
  list(msg = paste0("The message is: '", msg, "'"))
}
