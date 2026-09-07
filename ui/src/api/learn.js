import request from './request'

// 生成定制冲刺计划（AI 生成，耗时较长）
export function generateSprintPlan(data) {
  return request.post('/learn/plan', data)
}

// 获取当前冲刺计划（未定制时 data 为 null）
export function getSprintPlan() {
  return request.get('/learn/plan')
}

// 勾选/取消任务完成状态
export function updateSprintTask(data) {
  return request.put('/learn/plan/task', data)
}

// 删除当前冲刺计划
export function deleteSprintPlan() {
  return request.delete('/learn/plan')
}
