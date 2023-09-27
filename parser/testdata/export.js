export let foo = bar
export const foo = bar
export function foobar() {}

export {foo, bar}
export {foo as superfoo, bar}

import * as foobar from "foobar"
import {foo, bar} from "foobar"
import {foo as superfoo, bar} from "foobar"
import "foobar"