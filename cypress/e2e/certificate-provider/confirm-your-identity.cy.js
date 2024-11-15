describe('confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=confirmYourDetails');

        cy.contains('li', "Confirm your identity")
            .should('contain', 'Not started')
            .find('a')
            .click();
    })

    it('can see details of a successful ID check', () => {
        cy.contains('button', 'Continue').click()
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
        cy.contains('button', 'Continue').click()
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
        cy.contains('button', 'Continue').click()
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

    it('can go to the post office ', () => {
        cy.url().should('contain', '/confirm-your-identity');
        cy.contains('button', 'Continue').click();

        cy.go(-2);
        cy.contains('li', "Confirm your identity")
            .should('contain', 'In progress')
            .find('a')
            .click();

        cy.url().should('contain', '/how-will-you-confirm-your-identity');
        cy.checkA11yApp();
        cy.contains('label', 'I will confirm my identity at a Post Office').click();
        cy.contains('button', 'Continue').click();

        cy.contains('li', "Confirm your identity")
            .should('contain', 'Pending')
            .find('a')
            .click();
    });
})
