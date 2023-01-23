describe('Certificate provider task', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/task-list&withAttorney=1');
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

        cy.visitLpa('/task-list');

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
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type('07535111111');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'Online and by email').click();
        cy.get('#f-email').type('someone@example.com');
        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Solicitor').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/check-your-lpa');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.visitLpa('/task-list')
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
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-mobile').type('07535111111');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-would-certificate-provider-prefer-to-carry-out-their-role');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('label', 'Using paper forms').click();
        cy.contains('button', 'Continue').click()

        cy.url().should('contain', '/certificate-provider-address');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-select-address').select('2 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-address-line-1').should('have.value', '2 RICHMOND PLACE');
        cy.get('#f-address-line-2').should('have.value', '');
        cy.get('#f-address-line-3').should('have.value', '');
        cy.get('#f-address-town').should('have.value', 'BIRMINGHAM');
        cy.get('#f-address-postcode').should('have.value', 'B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-do-you-know-your-certificate-provider');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How do you know John Doe, your certificate provider?');
        cy.contains('label', 'Friend').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/how-long-have-you-known-certificate-provider');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.contains('How long have you known John Doe?');
        cy.contains('label', '2 years or more').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-to-notify-people');
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false }, 'aria-allowed-attr': { enabled: false } } });

        cy.visitLpa('/task-list')
        cy.contains('li', "Choose your certificate provider")
            .should('contain', 'Completed');
    });
});
