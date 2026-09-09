const formatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
})

export function formatMoney(cents: number): string {
  return formatter.format(cents / 100)
}

export function centsToInput(cents: number): string {
  return (cents / 100).toFixed(2)
}

export function centsFromInput(input: string): number | null {
  const dollars = evaluateFormula(input)
  if (dollars === null || dollars < 0) return null
  return toCents(dollars)
}

export function signedCentsFromInput(input: string): number | null {
  const dollars = evaluateFormula(input)
  if (dollars === null) return null
  return toCents(dollars)
}

function toCents(dollars: number): number | null {
  const cents = Math.round(dollars * 100)
  if (!Number.isSafeInteger(cents)) return null
  return cents
}

type FormulaOp = '+' | '-' | '*' | '/'

type FormulaToken =
  | { type: 'num'; value: number }
  | { type: 'op'; value: FormulaOp }
  | { type: 'lparen' }
  | { type: 'rparen' }

function tokenizeFormula(input: string): FormulaToken[] | null {
  const tokens: FormulaToken[] = []
  let i = 0
  while (i < input.length) {
    const ch = input[i]
    if (ch === ' ' || ch === '\t' || ch === '\n' || ch === '\r') {
      i++
      continue
    }
    if ((ch >= '0' && ch <= '9') || ch === '.') {
      let j = i
      let dots = 0
      while (
        j < input.length &&
        ((input[j] >= '0' && input[j] <= '9') || input[j] === '.')
      ) {
        if (input[j] === '.') dots++
        j++
      }
      if (dots > 1) return null
      const value = Number(input.slice(i, j))
      if (!Number.isFinite(value)) return null
      tokens.push({ type: 'num', value })
      i = j
      continue
    }
    if (ch === '+' || ch === '-' || ch === '*' || ch === '/') {
      tokens.push({ type: 'op', value: ch })
      i++
      continue
    }
    if (ch === '(') {
      tokens.push({ type: 'lparen' })
      i++
      continue
    }
    if (ch === ')') {
      tokens.push({ type: 'rparen' })
      i++
      continue
    }
    return null
  }
  return tokens
}

function parseFormulaSum(tokens: FormulaToken[], pos: { i: number }): number | null {
  let left = parseFormulaProduct(tokens, pos)
  if (left === null) return null
  while (pos.i < tokens.length) {
    const token = tokens[pos.i]
    if (token.type !== 'op' || (token.value !== '+' && token.value !== '-')) break
    pos.i++
    const right = parseFormulaProduct(tokens, pos)
    if (right === null) return null
    left = token.value === '+' ? left + right : left - right
  }
  return left
}

function parseFormulaProduct(tokens: FormulaToken[], pos: { i: number }): number | null {
  let left = parseFormulaFactor(tokens, pos)
  if (left === null) return null
  while (pos.i < tokens.length) {
    const token = tokens[pos.i]
    if (token.type !== 'op' || (token.value !== '*' && token.value !== '/')) break
    pos.i++
    const right = parseFormulaFactor(tokens, pos)
    if (right === null) return null
    if (token.value === '*') {
      left = left * right
    } else {
      if (right === 0) return null
      left = left / right
    }
  }
  return left
}

function parseFormulaFactor(tokens: FormulaToken[], pos: { i: number }): number | null {
  const token = tokens[pos.i]
  if (token === undefined) return null
  if (token.type === 'num') {
    pos.i++
    return token.value
  }
  if (token.type === 'op' && (token.value === '-' || token.value === '+')) {
    pos.i++
    const value = parseFormulaFactor(tokens, pos)
    if (value === null) return null
    return token.value === '-' ? -value : value
  }
  if (token.type === 'lparen') {
    pos.i++
    const value = parseFormulaSum(tokens, pos)
    if (value === null) return null
    const closing = tokens[pos.i]
    if (closing === undefined || closing.type !== 'rparen') return null
    pos.i++
    return value
  }
  return null
}

function evaluateFormula(input: string): number | null {
  const trimmed = input.trim()
  if (trimmed === '') return null
  const tokens = tokenizeFormula(trimmed)
  if (!tokens || tokens.length === 0) return null
  const pos = { i: 0 }
  const value = parseFormulaSum(tokens, pos)
  if (value === null || pos.i !== tokens.length) return null
  return Number.isFinite(value) ? value : null
}
