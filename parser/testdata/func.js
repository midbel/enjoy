function test() {
  return 0
}

function add(x, y) {
  return x + y
}

function add2(x = 1, y) {
  return x + y
}

function rest(x, y, ...rest) {
  return x + y + rest.length
}
