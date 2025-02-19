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
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Vivian Vaughn').click();
            cy.contains('button', 'Continue').click();
        });

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
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.visitLpa('/confirm-your-identity');

        cy.checkA11yApp();
        cy.contains('button', 'Continue').click();
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Charlie Cooper').click();
            cy.contains('button', 'Continue').click();
        });

        cy.url().should('contain', '/confirm-allowed-to-vouch');
        cy.checkA11yApp();
        cy.contains('Your confirmed identity details match someone');

        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.get('ul li:nth-child(3)').should('contain', 'Completed');
    });

    it('can update name', () => {
        cy.contains('button', 'Continue').click();
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Custom').click();
            cy.get('[name=first-names]').type('John');
            cy.get('[name=last-name]').type('Johnson');
            cy.get('[name=day]').type('2');
            cy.get('[name=month]').type('1');
            cy.get('[name=year]').type('1990');
            cy.contains('button', 'Continue').click();
        });

        cy.contains('a', 'Continue').click();
        cy.get('ul li:nth-child(3)').should('contain', 'Completed');
    });

    it('can update name to related', () => {
        cy.contains('button', 'Continue').click();
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Sam Smith').click();
            cy.contains('button', 'Continue').click();
        });

        cy.url().should('contain', '/confirm-allowed-to-vouch');
        cy.contains('label', 'Yes').click();
        cy.contains('button', 'Continue').click();
        cy.get('ul li:nth-child(3)').should('contain', 'Completed');
    });

    it('fails when fail one login', () => {
        cy.contains('button', 'Continue').click();
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Failed identity check').click();
            cy.contains('button', 'Continue').click();
        });

        cy.url().should('contain', '/voucher-unable-to-confirm-identity');
        cy.checkA11yApp();
        cy.contains('This means you cannot vouch for Sam Smith');
    });

    it('fails when related', () => {
        cy.contains('button', 'Continue').click();
        cy.origin('http://localhost:7012', () => {
            cy.contains('label', 'Sam Smith').click();
            cy.contains('button', 'Continue').click();
        });

        cy.url().should('contain', '/confirm-allowed-to-vouch');
        cy.contains('label', 'No').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/you-cannot-vouch-for-donor');
        cy.checkA11yApp();
        cy.contains('You have told us that you cannot vouch for Sam Smith');
    });

    it('can go to the post office ', () => {
        cy.url().should('contain', '/confirm-your-identity');

        cy.url().then(u => {
            cy.contains('button', 'Continue').click();
            cy.visit(u.split('/').slice(3, -1).join('/') + '/task-list');
        });

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
