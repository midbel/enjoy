let ok = false
let age
if (ok) {
  age = 10
} else {
  age = 0
}

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
} catch(error) {
  // pass
}

for (let i = 0; i < 100; i++) {
  age -= i
}
