import { ChatPage, CollapsibleOutputPage } from '../support/page-objects'

describe('P1-3: Long Content Collapsing - Comprehensive', () => {
  let chatPage: ChatPage
  let outputPage: CollapsibleOutputPage

  beforeEach(() => {
    chatPage = new ChatPage()
    outputPage = new CollapsibleOutputPage()
    chatPage.visit()
  })

  describe('TC01: Output over 50 lines shows collapse', () => {
    it('should display collapse button for long output', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.getCollapsibleOutput().should('exist')
      outputPage.getCollapseToggle().should('be.visible')
      outputPage.isCollapsed()
    })

    it('should show omitted content indicator', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.hasOmittedContent()
    })

    it('should display line count in toggle button', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.getCollapseToggle().should('contain', '展开全部')
      outputPage.getCollapseToggle().should('contain', '行')
    })
  })

  describe('TC02: Expand button shows full content', () => {
    it('should expand content when clicking expand button', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.isExpanded()
      outputPage.hasNoOmittedContent()
    })

    it('should display all lines after expansion', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.getOutputContent().should('not.contain', '... 省略')
    })

    it('should update button text to collapse', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.getCollapseToggle().should('contain', '收起')
    })
  })

  describe('TC03: Collapse button restores collapsed state', () => {
    it('should collapse content when clicking collapse button', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.isExpanded()

      outputPage.toggleCollapse()
      outputPage.isCollapsed()
      outputPage.hasOmittedContent()
    })

    it('should restore omitted content indicator', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.hasNoOmittedContent()

      outputPage.toggleCollapse()
      outputPage.hasOmittedContent()
    })

    it('should update button text back to expand', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.getCollapseToggle().should('contain', '收起')

      outputPage.toggleCollapse()
      outputPage.getCollapseToggle().should('contain', '展开全部')
    })
  })

  describe('TC04: Short content does not show collapse button', () => {
    it('should not display collapse button for short output', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      outputPage.getCollapsibleOutput().should('exist')
      outputPage.getCollapseToggle().should('not.exist')
    })

    it('should display full content without collapse', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      outputPage.getOutputContent().should('not.contain', '... 省略')
    })
  })

  describe('Multiple outputs with different lengths', () => {
    it('should handle multiple outputs correctly', () => {
      chatPage.selectHosts(['host-001'])

      chatPage.sendMessage('查看系统负载')
      cy.get('[data-testid="collapsible-output"]').eq(0).within(() => {
        cy.get('[data-testid="collapse-toggle"]').should('not.exist')
      })

      chatPage.sendMessage('查看完整日志')
      cy.get('[data-testid="collapsible-output"]').eq(1).within(() => {
        cy.get('[data-testid="collapse-toggle"]').should('be.visible')
      })
    })
  })

  describe('Collapse state persistence', () => {
    it('should maintain collapse state when scrolling', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看完整日志')

      outputPage.toggleCollapse()
      outputPage.isExpanded()

      cy.get('[data-testid="message-list"]').scrollTo('top')
      outputPage.isExpanded()
    })
  })
})
