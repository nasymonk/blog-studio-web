export function countWords(text: string): number {
  const chinese = (text.match(/[一-龥]/g) || []).length
  const english = (text.match(/[a-zA-Z]+/g) || []).length
  return chinese + english
}
