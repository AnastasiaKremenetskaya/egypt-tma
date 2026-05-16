export type Phase = 'lobby' | 'question' | 'voting' | 'seth' | 'finished'

export interface PlayerView {
  user_id: number
  username: string
  title: string
  score: number
}

export interface QuestionView {
  id: string
  type: 'maat' | 'seth'
  text: string
  options?: string[]
}

export interface AnswerView {
  type: 'text' | 'voice'
  text?: string
}

export interface RoomState {
  code: string
  admin_id: number
  phase: Phase
  players: PlayerView[]
  active_idx: number
  round: number
  question?: QuestionView
  answer?: AnswerView
  vote_trust: number
  vote_lie: number
  voted_ids: number[]
  seth_answered_ids: number[]
  phase_deadline: string // ISO date string
}

export interface WSMessage {
  type: 'state'
  state: RoomState
}
