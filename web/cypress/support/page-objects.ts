// Page Object Model for Chat Window
export class ChatPage {
  visit() {
    cy.visit('/')
  }

  getInputBox() {
    return cy.get('[data-testid="input-box"]')
  }

  getSendButton() {
    return cy.get('[data-testid="send-button"]')
  }

  getMessageList() {
    return cy.get('[data-testid="message-list"]')
  }

  getHostSelector() {
    return cy.get('[data-testid="host-selector"]')
  }

  getHostOption(hostId: string) {
    return cy.get(`[data-testid="host-option-${hostId}"]`)
  }

  getSessionList() {
    return cy.get('[data-testid="session-list"]')
  }

  getSessionItem(index: number) {
    return cy.get('[data-testid="session-item"]').eq(index)
  }

  getNewSessionButton() {
    return cy.get('[data-testid="new-session-button"]')
  }

  getRightPanel() {
    return cy.get('[data-testid="right-panel"]')
  }

  getCollapseButton() {
    return cy.get('[data-testid="collapse-button"]')
  }

  getResultArea() {
    return cy.get('[data-testid="result-area"]')
  }

  sendMessage(message: string) {
    this.getInputBox().type(message)
    this.getSendButton().click()
    cy.wait(500)
  }

  selectHosts(hostIds: string[]) {
    this.getHostSelector().click()
    hostIds.forEach(hostId => {
      this.getHostOption(hostId).click()
    })
    cy.get('body').click(0, 0)
  }

  getLastMessage() {
    return this.getMessageList().find('[data-testid="message"]').last()
  }

  getHostTags() {
    return cy.get('[data-testid="host-tag"]')
  }

  getSessionHosts() {
    return cy.get('[data-testid="session-hosts"]')
  }
}

// Page Object Model for Danger Confirmation Dialog
export class DangerConfirmPage {
  getDialog() {
    return cy.get('[data-testid="danger-confirm-dialog"]')
  }

  getDialogTitle() {
    return cy.get('[data-testid="dialog-title"]')
  }

  getRiskLevel() {
    return cy.get('[data-testid="risk-level"]')
  }

  getConfirmCheckbox() {
    return cy.get('[data-testid="danger-checkbox"]')
  }

  getConfirmInput() {
    return cy.get('[data-testid="confirm-input"]')
  }

  getConfirmButton() {
    return cy.get('[data-testid="confirm-button"]')
  }

  getCancelButton() {
    return cy.get('[data-testid="cancel-button"]')
  }

  getAffectedHosts() {
    return cy.get('[data-testid="affected-hosts"]')
  }

  confirm(confirmText?: string) {
    this.getConfirmCheckbox().click()
    if (confirmText) {
      this.getConfirmInput().type(confirmText)
    }
    this.getConfirmButton().click()
  }

  cancel() {
    this.getCancelButton().click()
  }

  isVisible() {
    return this.getDialog().should('be.visible')
  }

  isHidden() {
    return this.getDialog().should('not.exist')
  }
}

// Page Object Model for Collapsible Output
export class CollapsibleOutputPage {
  getCollapsibleOutput() {
    return cy.get('[data-testid="collapsible-output"]')
  }

  getCollapseToggle() {
    return cy.get('[data-testid="collapse-toggle"]')
  }

  getOutputContent() {
    return cy.get('[data-testid="output-content"]')
  }

  getOutputHeader() {
    return cy.get('[data-testid="output-header"]')
  }

  toggleCollapse() {
    this.getCollapseToggle().click()
  }

  isExpanded() {
    return this.getCollapseToggle().should('contain', '收起')
  }

  isCollapsed() {
    return this.getCollapseToggle().should('contain', '展开全部')
  }

  hasOmittedContent() {
    return this.getOutputContent().should('contain', '... 省略')
  }

  hasNoOmittedContent() {
    return this.getOutputContent().should('not.contain', '... 省略')
  }
}
