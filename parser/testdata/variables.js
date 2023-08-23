// assignment
let lang;
lang = "javascript"

let who = "foobar"
const who = "foobar"

let arr = []
let obj = {}

arr = [1, true, [0, false, Math.Pi], {user: who, pass: who}]
obj.user = who
obj.pass = "*****"

// destructuring arguments
const [x = 1, y = 2, ...rest] = [,,,"foo", "bar"]
const {user, pass: password = "******", ...rest} = obj

// operators
let str = "foobar" + -101
let cent = ((1 / 100) + (100*10)) % 100 ** 1

cent /= 10 << 2

1 === "1"
who = who || "nobody" || undefined

null == null == undefined
