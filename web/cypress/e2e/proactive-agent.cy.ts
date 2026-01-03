// E2E 测试：主动消息和健康仪表盘功能
describe('Proactive Agent Features', () => {
  beforeEach(() => {
    cy.visit('/')
  })

  describe('主动消息展示', () => {
    it('应该在聊天流中展示主动消息', () => {
      // 模拟 WebSocket 主动消息
      cy.window().then((win) => {
        const proactiveMessage = {
          type: 'proactive_message',
          data: {
            id: 'test-proactive-1',
            type: 'morning_report',
            title: '早安报告',
            content: '系统运行正常，所有主机健康分均在 90 分以上。',
            priority: 'low',
            created_at: Math.floor(Date.now() / 1000)
          }
        }
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: proactiveMessage }))
      })

      // 验证主动消息卡片显示
      cy.get('.proactive-message').should('exist')
      cy.get('.proactive-message .title-text').should('contain', '早安报告')
      cy.get('.proactive-message .monitoring-badge').should('contain', '监控中')
    })

    it('应该根据优先级显示不同样式', () => {
      const priorities = ['critical', 'high', 'medium', 'low']

      priorities.forEach((priority) => {
        cy.window().then((win) => {
          const message = {
            type: 'proactive_message',
            data: {
              id: `test-${priority}`,
              type: 'anomaly_alert',
              title: `${priority} 级别告警`,
              content: '测试内容',
              priority,
              created_at: Math.floor(Date.now() / 1000)
            }
          }
          win.dispatchEvent(new CustomEvent('websocket-message', { detail: message }))
        })
      })

      // 验证不同优先级的样式类
      cy.get('.proactive-message.priority-critical').should('exist')
      cy.get('.proactive-message.priority-high').should('exist')
    })

    it('应该显示快捷操作按钮', () => {
      cy.window().then((win) => {
        const message = {
          type: 'proactive_message',
          data: {
            id: 'test-actions',
            type: 'anomaly_alert',
            title: '异常告警',
            content: '发现异常',
            priority: 'high',
            actions: [
              { id: '1', label: '查看详情' },
              { id: '2', label: '运行诊断', dangerous: true }
            ],
            created_at: Math.floor(Date.now() / 1000)
          }
        }
        win.dispatchEvent(new CustomEvent('websocket-message', { detail: message }))
      })

      cy.get('.proactive-actions .el-button').should('have.length', 2)
      cy.get('.proactive-actions .el-button').first().should('contain', '查看详情')
    })
  })

  describe('健康仪表盘', () => {
    beforeEach(() => {
      // 导航到包含健康仪表盘的页面
      cy.visit('/dashboard')
    })

    it('应该显示整体健康分', () => {
      cy.get('.health-dashboard').should('exist')
      cy.get('.overall-health').should('exist')
      cy.get('.score-value').should('exist')
      cy.get('.score-label').should('contain', '健康分')
    })

    it('应该能够触发健康检查', () => {
      cy.intercept('POST', '/api/health/compare', {
        statusCode: 200,
        body: {
          hosts: [
            { host_id: '1', host_name: 'server-1', score: 95, score_level: 'excellent', cpu: 30, memory: 45, disk: 60, status: 'healthy', issue_count: 0 },
            { host_id: '2', host_name: 'server-2', score: 75, score_level: 'good', cpu: 70, memory: 65, disk: 55, status: 'warning', issue_count: 1 }
          ],
          highest_cpu: 'server-2',
          highest_mem: 'server-2',
          highest_disk: 'server-1',
          anomalies: [],
          avg_score: 85
        }
      }).as('healthCheck')

      cy.get('.section-header .el-button').click()
      cy.wait('@healthCheck')

      // 验证表格数据
      cy.get('.el-table__body-wrapper').should('contain', 'server-1')
      cy.get('.el-table__body-wrapper').should('contain', 'server-2')
    })

    it('应该显示异常主机告警', () => {
      cy.intercept('POST', '/api/health/compare', {
        statusCode: 200,
        body: {
          hosts: [
            { host_id: '1', host_name: 'server-1', score: 45, score_level: 'critical', cpu: 95, memory: 90, disk: 85, status: 'critical', issue_count: 3 }
          ],
          anomalies: ['server-1'],
          avg_score: 45
        }
      }).as('healthCheck')

      cy.get('.section-header .el-button').click()
      cy.wait('@healthCheck')

      cy.get('.anomaly-alert').should('exist')
      cy.get('.anomaly-alert').should('contain', '1 台异常主机')
    })

    it('应该根据分数显示正确的等级颜色', () => {
      cy.intercept('GET', '/api/health/overall', {
        statusCode: 200,
        body: { score: 95, level: 'excellent' }
      }).as('overallHealth')

      cy.wait('@overallHealth')
      cy.get('.score-circle').should('have.class', 'excellent')
    })
  })

  describe('健康分计算', () => {
    it('应该正确计算并显示健康分', () => {
      cy.intercept('POST', '/api/health/check', {
        statusCode: 200,
        body: {
          host_id: '1',
          host_name: 'test-server',
          status: 'healthy',
          score: 88,
          score_level: 'good',
          metrics: {
            cpu_usage: 45,
            memory_usage: 55,
            disk_usage: 60
          },
          issues: [],
          suggestions: []
        }
      }).as('healthCheck')

      // 触发健康检查
      cy.request('POST', '/api/health/check', { host_id: '1' }).then((response) => {
        expect(response.body.score).to.be.within(0, 100)
        expect(response.body.score_level).to.be.oneOf(['excellent', 'good', 'warning', 'critical'])
      })
    })
  })

  describe('横向对比功能', () => {
    it('应该能够对比多个主机', () => {
      cy.intercept('POST', '/api/health/compare', {
        statusCode: 200,
        body: {
          hosts: [
            { host_id: '1', host_name: 'server-1', score: 90, cpu: 40, memory: 50, disk: 60 },
            { host_id: '2', host_name: 'server-2', score: 70, cpu: 80, memory: 75, disk: 65 },
            { host_id: '3', host_name: 'server-3', score: 85, cpu: 55, memory: 60, disk: 70 }
          ],
          highest_cpu: 'server-2',
          highest_mem: 'server-2',
          highest_disk: 'server-3',
          anomalies: [],
          avg_score: 82
        }
      }).as('compareHosts')

      cy.request('POST', '/api/health/compare', { host_ids: ['1', '2', '3'] }).then((response) => {
        expect(response.body.hosts).to.have.length(3)
        expect(response.body.highest_cpu).to.equal('server-2')
        expect(response.body.avg_score).to.equal(82)
      })
    })
  })
})
