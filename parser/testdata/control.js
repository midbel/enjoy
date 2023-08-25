let ok = false
let age
if (ok) {
  age = 10
} else {
  age = 0
}

age = ok ? 100 : 50

while (age > 0) {
  --age
  if (age % 2 == 0) {
    continue
  }
}

do {
  ++age
  if (age == 7) {
    break
  }
} while (age < 10)

try {
  const i = 1
  i / 0
} catch(err) {
  throw err
} finally {
  console.log("run finalizer")
}

const arr = [100, 101, 200, true, false, "foo", "bar"]
let i
for (i = 0; i < 100; i += 1) {
  console.log(i)
}

for (;;) {
  console.log("eternal loop")
}

for (; i < 100; i += 5) {
  console.log(i)
}

for (; i <= 100; ) {
  console.log(i)
}

for (let i = 0; i < 100; i += 1) {
  console.log(i)
}

const object = { a: 1, b: 2, c: 3 };

const obj = {
  name: "foobar",
  pass: "tmp123",
}
for (const prop in obj) {
  console.log(`${property}: ${object[prop]}`);
}

for (const el of arr) {
  console.log(el);
}

for (const { user, pass } of users) {
  const weak = isWeakPassword(pass)
  if (!weak) {
    continue
  }
  console.log(`${user}: weak password ${pass}`)
}
