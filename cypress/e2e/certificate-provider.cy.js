describe('Certificate provider task', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list');
    });

    it('can be done later', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'I will do this later').click();

        cy.url().should('contain', '/task-list');
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started');
    });

    it('can be left unfinished', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/certificate-provider-details');

        cy.visit('/task-list');

        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'In progress');
    });

    it('can be a professional', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/certificate-provider-details');

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type('what');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Solicitor').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/task-list');
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });

    it('can be a lay person', () => {
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/certificate-provider-details');

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-email').type('what');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Friend').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/how-long-have-you-known-certificate-provider');

        cy.contains('How long have you known John Doe?');
        cy.contains('label', '2 years or more').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/task-list');
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });
});
