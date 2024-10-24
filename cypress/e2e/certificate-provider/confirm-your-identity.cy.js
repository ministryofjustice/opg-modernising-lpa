describe('confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=confirmYourDetails');

        cy.url().should('contain', '/task-list');
        cy.checkA11yApp();

        cy.contains('li', 'Confirm your identity').should('contain', 'Not started').click();

        cy.url().should('contain', '/confirm-your-identity');
        cy.checkA11yApp();

        cy.contains('a', 'Continue').click()
    })

    it('can see details of a successful ID check', () => {
        cy.get('[name="user"]').check('certificate-provider', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/one-login-identity-details');
        cy.checkA11yApp();

        cy.contains('Charlie')
        cy.contains('Cooper')
        cy.contains('2 January 1990')

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/read-the-lpa');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');

        cy.contains('li', 'Confirm your identity').should('contain', 'Completed').click();

        cy.url().should('contain', '/read-the-lpa');
    })

    it('can see next steps when failing an ID check', () => {
        cy.get('[name="return-code"]').check('T', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/unable-to-confirm-identity');
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/read-the-lpa');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your identity').should('contain', 'Completed').click();

        cy.url().should('contain', '/read-the-lpa');
    })

    it('can see next steps when has insufficient evidence for ID', () => {
        cy.get('[name="return-code"]').check('X', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/unable-to-confirm-identity');
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/read-the-lpa');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your identity').should('contain', 'Completed').click();

        cy.url().should('contain', '/read-the-lpa');
    })
})
