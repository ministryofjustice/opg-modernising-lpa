describe('Confirm your identity', () => {
    beforeEach(() => {
        cy.visit('/fixtures/voucher?redirect=/task-list&progress=verifyDonorDetails');

        cy.contains('li', "Confirm your identity")
            .should('contain', 'Not started')
            .find('a')
            .click();
    });

    it('can be confirmed', () => {
        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();
        cy.contains('label', 'Vivian Vaughn').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/one-login-identity-details');
        cy.checkA11yApp();
        cy.contains('a', 'Continue').click();

        cy.get('.govuk-task-list li:nth-child(3)').should('contain', 'Completed');
        cy.contains('a', 'Confirm your identity').click();

        cy.url().should('contain', '/one-login-identity-details');
        cy.contains('a', 'Continue').click();

        cy.contains('a', 'Confirm your name').click();
        cy.contains('a', 'Change').should('not.exist');

        cy.contains('a', 'Manage your LPAs').click();
        cy.contains('I’m vouching for someone');
    });

    it('warns when matches another actor', () => {
        cy.visitLpa('/your-name');
        cy.get('#f-first-names').clear().type('Charlie');
        cy.get('#f-last-name').clear().type('Cooper');
        cy.contains('button', 'Save and continue').click();
        cy.contains('button', 'Continue').click();
        cy.visitLpa('/confirm-your-identity');

        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();
        cy.contains('label', 'Charlie Cooper').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-allowed-to-vouch');
        cy.checkA11yApp();
        cy.contains('Your confirmed identity details match someone');

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.get('ul li:nth-child(3)').should('contain', 'Completed');
    });

    it('can fail', () => {
        cy.contains('button', 'Continue').click();
        cy.contains('label', 'Sam Smith').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/voucher-unable-to-confirm-identity');
    });

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
});
