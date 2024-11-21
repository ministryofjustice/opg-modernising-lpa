describe('confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/certificate-provider?redirect=/task-list&progress=confirmYourDetails');

        cy.contains('li', "Confirm your identity")
            .should('contain', 'Not started')
            .find('a')
            .click();
    })

    it('can see details when successful', () => {
        cy.contains('button', 'Continue').click()
        cy.get('[name="user"]').check('certificate-provider', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/identity-details');
        cy.checkA11yApp();

        cy.contains('Charlie')
        cy.contains('Cooper')
        cy.contains('2 January 1990')

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your identity').should('contain', 'Completed').click();

        cy.url().should('contain', '/identity-details');
        cy.contains('You have successfully confirmed your identity');
    })

    it('can see details when not matched', () => {
        cy.contains('button', 'Continue').click()
        cy.get('[name="user"]').check('donor', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/identity-details');
        cy.checkA11yApp();

        cy.contains('Charlie')
        cy.contains('Cooper')
        cy.contains('2 January 1990')

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your identity').should('contain', 'Pending').click();

        cy.url().should('contain', '/identity-details');
        cy.contains('Some of the details on the LPA do not match');
    })

    it('can see next steps when failing', () => {
        cy.contains('button', 'Continue').click()
        cy.get('[name="return-code"]').check('T', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/identity-details');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your identity').should('contain', 'Completed').click();

        cy.url().should('contain', '/identity-details');
        cy.contains('You were not able to confirm your identity');
    })

    it('can see next steps when has insufficient evidence', () => {
        cy.contains('button', 'Continue').click()
        cy.get('[name="return-code"]').check('X', { force: true })

        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/identity-details');
        cy.checkA11yApp();

        cy.contains('a', 'Return to task list').click()

        cy.url().should('contain', '/task-list');
        cy.contains('li', 'Confirm your identity').should('contain', 'Completed').click();

        cy.url().should('contain', '/identity-details');
        cy.contains('You were not able to confirm your identity');
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

        cy.url().should('contain', '/completing-your-identity-confirmation');
    });
})
