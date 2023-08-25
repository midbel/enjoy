const arr_1 = []
arr_1.map(x => 2 ** 2)
arr_1.map(x => {
  x <<= 7
  return x
})

arr_1.map(x => ({
  num: x,
  label: `item-${x}`,
}))

arrow = (x1, y1, ...rest) => 42

arrow = (a = 42, b) => a

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
