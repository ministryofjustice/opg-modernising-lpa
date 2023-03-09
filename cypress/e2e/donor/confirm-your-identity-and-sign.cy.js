describe('Confirm your identity and sign', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/your-details&withIncompleteAttorneys=1&withCP=1&paymentComplete=1');
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();
        cy.visitLpa('/task-list');
    });

    it('can be completed', () => {
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.url().should('contain', '/how-to-confirm-your-identity-and-sign');
        cy.checkA11yApp();

        cy.contains('h1', 'How to confirm your identity and sign the LPA');
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/what-youll-need-to-confirm-your-identity');
        cy.checkA11yApp();

        cy.contains('h1', "What you’ll need to confirm your identity");
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/select-your-identity-options');
        cy.checkA11yApp();

        cy.contains('label', 'I do not have either of these types of accounts').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/select-your-identity-options-1');
        cy.checkA11yApp();

        cy.contains('label', 'Your passport').click();
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/your-chosen-identity-options');
        cy.checkA11yApp();
        
        cy.contains('passport');
        cy.contains('button', 'Continue').click();
        cy.contains('button', 'Continue').click();
        
        cy.url().should('contain', '/read-your-lpa');
        cy.checkA11yApp();

        cy.contains('h2', "LPA decisions");
        cy.contains('h2', "People named on the LPA");
        cy.contains('h3', "Donor");
        cy.contains('h3', "Attorneys");
        cy.contains('h3', "Replacement attorney");
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/your-legal-rights-and-responsibilities');
        cy.checkA11yApp();
        cy.contains('a', 'Continue to signing page').click();

        cy.url().should('contain', '/sign-your-lpa');
        cy.checkA11yApp();

        cy.contains('h1', "Sign your LPA");
        cy.contains('label', 'I want to sign this LPA').click();
        cy.contains('label', 'I want to apply to register this LPA').click();
        cy.contains('button', 'Submit my signature').click();

        cy.url().should('contain', '/witnessing-your-signature');
        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/witnessing-as-certificate-provider');
        cy.checkA11yApp();

        cy.contains('h1', "Witnessing as the certificate provider");
        cy.get('#f-witness-code').type('1234');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/you-have-submitted-your-lpa');
        cy.checkA11yApp();

        cy.contains('h1', "You’ve submitted your LPA");
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/dashboard');
        cy.checkA11yApp();
    });

    it('can be restarted', () => {
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.contains('a', 'Continue').click();
        cy.contains('a', 'Continue').click();
        cy.contains('label', 'Your GOV.UK One Login Identity').click();
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/task-list');

        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'In progress')
            .find('a')
            .click();

        cy.contains('a', 'Continue').click();
        cy.contains('a', 'Continue').click();
        cy.contains('button', 'Continue').click();
        cy.contains('Your GOV.UK One Login Identity');
    });

    it('errors when not signed', () => {
        cy.visitLpa('/sign-your-lpa');

        cy.contains('button', 'Submit my signature').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('You must select both boxes to sign and apply to register your LPA');
        });

        cy.contains('.moj-ticket-panel  .govuk-error-message', 'You must select both boxes to sign and apply to register your LPA');
    });

    it('errors when not witnessed', () => {
        cy.visitLpa('/id/passport');
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/witnessing-your-signature');
        cy.contains('button', 'Continue').click();

        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter the code we sent to the certificate provider');
        });

        cy.contains('.moj-ticket-panel .govuk-error-message', 'Enter the code we sent to the certificate provider');

        cy.get('#f-witness-code').type('123');
        cy.contains('button', 'Continue').click();

        cy.contains('.moj-ticket-panel .govuk-error-message', 'The code we sent to the certificate provider must be 4 characters');

        cy.get('#f-witness-code').type('45');
        cy.contains('button', 'Continue').click();

        cy.contains('.moj-ticket-panel .govuk-error-message', 'The code we sent to the certificate provider must be 4 characters');
    });
});
