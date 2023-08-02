describe('Enter date of birth', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/enter-date-of-birth&asCertificateProvider=1&loginAs=certificate-provider');
    });

    it('can be completed', () => {
        cy.checkA11yApp();

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-your-details');
    });

    it('can be over 100', () => {
        cy.checkA11yApp();

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1900');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/enter-date-of-birth');
        cy.contains('By saving this section, you confirm that the person is more than 100 years old')

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/confirm-your-details');
    });

    it('errors when empty', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter date of birth');
        });

        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
    });

    it('errors when invalid dates of birth', () => {
        cy.get('#f-date-of-birth').type('1');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must include a month and year');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('2222');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be in the past');

        cy.get('#f-date-of-birth').type('not');
        cy.get('#f-date-of-birth-month').type('valid');
        cy.get('#f-date-of-birth-year').clear().type('values');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be a real date');
    });

    it('errors when not over 18', () => {
        const lastYear = (new Date(new Date().setFullYear(new Date().getFullYear() - 1))).getFullYear()

        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type(lastYear.toString());
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('You’ve entered a date of birth that means you are under 18, you must be 18 to be a certificate provider');
        });

        cy.contains('#date-of-birth-hint + .govuk-error-message', 'You’ve entered a date of birth that means you are under 18, you must be 18 to be a certificate provider');
    });
});
