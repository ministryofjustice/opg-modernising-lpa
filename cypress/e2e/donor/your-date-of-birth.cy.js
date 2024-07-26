describe('Your date of birth', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-date-of-birth');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-date-of-birth').type('1');
            cy.get('#f-date-of-birth-month').type('2');
            cy.get('#f-date-of-birth-year').type('1990');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-address');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter date of birth');
            });

            cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
        });

        it('errors when invalid dates of birth', () => {
            cy.get('#f-date-of-birth').type('1');
            cy.contains('button', 'Save and continue').click();
            cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must include a month and year');

            cy.get('#f-date-of-birth-month').type('2');
            cy.get('#f-date-of-birth-year').type('2222');
            cy.contains('button', 'Save and continue').click();
            cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be in the past');

            cy.get('#f-date-of-birth-month').type('2');
            cy.get('#f-date-of-birth-year').clear().type('1990');
            cy.contains('button', 'Save and continue').click();
            cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be a real date');
        });

        it('permanently warns when date of birth is under 18', () => {
            cy.get('#f-date-of-birth').type('1');
            cy.get('#f-date-of-birth-month').type('2');
            cy.get('#f-date-of-birth-year').type(new Date().getFullYear() - 1);
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');

            cy.contains('You are under 18. By continuing, you understand that you must be at least 18 years old on the date you sign the LPA, or it will be rejected.');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-address');

            cy.visitLpa("/your-date-of-birth")
            cy.url().should('contain', '/your-date-of-birth');

            cy.contains('You are under 18. By continuing, you understand that you must be at least 18 years old on the date you sign the LPA, or it will be rejected.');
        });

        it('warns when date of birth is over 100', () => {
            cy.get('#f-date-of-birth').type('1');
            cy.get('#f-date-of-birth-month').type('2');
            cy.get('#f-date-of-birth-year').type('1900');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');

            cy.contains('By continuing, you confirm that this person is more than 100 years old. If not, please change their date of birth.');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-address');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-date-of-birth&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-details');
        });
    });
});
