import { ChatPage } from '../support/page-objects'

describe('P0-2: Host Tags Display - Comprehensive', () => {
  let chatPage: ChatPage

  beforeEach(() => {
    chatPage = new ChatPage()
    chatPage.visit()
  })

  describe('TC01: Single host execution shows host tag', () => {
    it('should display single host tag', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().should('have.length', 1)
      chatPage.getHostTags().eq(0).should('contain', 'host-001')
    })

    it('should display host tag with correct styling', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(0).should('have.class', 'status-success')
    })
  })

  describe('TC02: Multiple hosts execution shows multiple tags', () => {
    it('should display 3 host tags', () => {
      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().should('have.length', 3)
      chatPage.getHostTags().eq(0).should('contain', 'host-001')
      chatPage.getHostTags().eq(1).should('contain', 'host-002')
      chatPage.getHostTags().eq(2).should('contain', 'host-003')
    })

    it('should display all tags in correct order', () => {
      chatPage.selectHosts(['host-003', 'host-001', 'host-002'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(0).should('contain', 'host-003')
      chatPage.getHostTags().eq(1).should('contain', 'host-001')
      chatPage.getHostTags().eq(2).should('contain', 'host-002')
    })
  })

  describe('TC03: Offline host shows red status', () => {
    it('should display failed host in red', () => {
      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(1).should('have.class', 'status-failed')
      chatPage.getHostTags().eq(1).should('have.css', 'background-color', 'rgb(245, 108, 108)')
    })

    it('should show error indicator for offline host', () => {
      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(1).should('contain', '离线')
    })

    it('should display success status for online hosts', () => {
      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(0).should('have.class', 'status-success')
      chatPage.getHostTags().eq(2).should('have.class', 'status-success')
    })
  })

  describe('TC04: Collapse button toggles result area', () => {
    it('should collapse result area when clicking collapse button', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getResultArea().should('be.visible')
      chatPage.getCollapseButton().click()
      chatPage.getResultArea().should('not.be.visible')
    })

    it('should expand result area when clicking expand button', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getCollapseButton().click()
      chatPage.getResultArea().should('not.be.visible')

      chatPage.getCollapseButton().click()
      chatPage.getResultArea().should('be.visible')
    })

    it('should toggle collapse button icon', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getCollapseButton().should('have.class', 'expanded')
      chatPage.getCollapseButton().click()
      chatPage.getCollapseButton().should('not.have.class', 'expanded')
    })
  })

  describe('Multiple messages with different host counts', () => {
    it('should display correct tags for each message', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')
      cy.get('[data-testid="host-tag"]').should('have.length', 1)

      chatPage.selectHosts(['host-001', 'host-002', 'host-003'])
      chatPage.sendMessage('查看进程列表')
      cy.get('[data-testid="host-tag"]').should('have.length.at.least', 3)
    })
  })

  describe('Host tag interactions', () => {
    it('should show host details on hover', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(0).trigger('mouseenter')
      cy.get('[data-testid="host-tooltip"]').should('be.visible')
    })

    it('should allow copying host name', () => {
      chatPage.selectHosts(['host-001'])
      chatPage.sendMessage('查看系统负载')

      chatPage.getHostTags().eq(0).rightclick()
      cy.get('[data-testid="copy-host-name"]').should('be.visible')
    })
  })
})
