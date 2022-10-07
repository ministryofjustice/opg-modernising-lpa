describe('Confirm your identity and sign', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list');
    });

    it('can be completed', () => {
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'A passport').click();
        cy.contains('label', 'A driving licence').click();
        cy.contains('label', 'A government gateway account').click();
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('using your passport and driving licence');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('confirmed with passport');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('confirmed with driving licence');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.url().should('contain', '/what-happens-when-signing');
        cy.contains('a', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.url().should('contain', '/how-to-sign');
        cy.contains('a', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.url().should('contain', '/read-your-lpa');

        cy.contains('h2', "LPA decisions");
        cy.contains('h2', "People named on the LPA");
        cy.contains('h3', "Donor");
        cy.contains('h3', "Attorneys");
        cy.contains('h3', "Replacement attorney");

        cy.get('#f-checked').check();
        cy.get('#f-confirm').check();
        cy.get('#f-signature').type('anything');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.url().should('contain', '/signing-confirmation');
        cy.contains("You've signed your LPA");
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Completed');
    });

    it('can be restarted', () => {
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.contains('label', 'A passport').click();
        cy.contains('label', 'A driving licence').click();
        cy.contains('label', 'A government gateway account').click();
        cy.contains('button', 'Continue').click();

        cy.visit('/task-list');

        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'In progress')
            .find('a')
            .click();

        cy.contains('button', 'Continue').click();
        cy.contains('using your passport and driving licence');
    });
});
