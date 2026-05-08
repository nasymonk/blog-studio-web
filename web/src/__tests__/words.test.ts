import { describe, it, expect } from 'vitest'
import { countWords } from '../utils/words'

describe('countWords', () => {
  it('counts Chinese characters individually', () => {
    expect(countWords('你好世界')).toBe(4)
  })

  it('counts English words as units', () => {
    expect(countWords('hello world')).toBe(2)
  })

  it('counts mixed Chinese and English', () => {
    // 学习 = 2, 进阶 = 2 Chinese chars; JavaScript = 1 English word → total 5
    expect(countWords('学习 JavaScript 进阶')).toBe(4 + 1)
  })

  it('ignores punctuation and numbers', () => {
    expect(countWords('123 ，。！')).toBe(0)
  })

  it('handles empty string', () => {
    expect(countWords('')).toBe(0)
  })

  it('handles text with only spaces', () => {
    expect(countWords('   ')).toBe(0)
  })
})
