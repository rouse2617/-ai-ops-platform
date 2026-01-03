import request from '@/utils/request'

export interface GitCommit {
  hash: string
  author: string
  email: string
  message: string
  timestamp: string
  changed_files: string[]
  repository: string
  branch: string
}

export interface SuspiciousFile {
  file_path: string
  commit_hash: string
  change_type: string
  score: number
  reason: string
  diff_url: string
}

export interface CodeCorrelation {
  alert_id: string
  alert_name: string
  service: string
  correlation_score: number
  message: string
  commits: GitCommit[]
  suspicious_files: SuspiciousFile[]
  time_window: {
    alert_time: string
    search_start: string
    search_end: string
  }
  timestamp: string
}

export function getDiff(repository: string, commitHash: string, filePath?: string) {
  return request({
    url: `/api/git/${repository}/diff/${commitHash}`,
    method: 'get',
    params: { file: filePath }
  })
}

export function getCommits(repository: string, startTime: string, endTime: string) {
  return request({
    url: '/api/git/commits',
    method: 'post',
    data: {
      repository,
      start_time: startTime,
      end_time: endTime
    }
  })
}
