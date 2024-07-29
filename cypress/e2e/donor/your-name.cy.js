describe('Your name', () => {
    describe('first time', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-name');
        });

        it('can be submitted', () => {
            cy.checkA11yApp();
            cy.contains('a', 'Return to task list').should('not.exist');

            cy.get('#f-first-names').type('John');
            cy.get('#f-last-name').type('Doe');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        it('errors when empty', () => {
            cy.contains('button', 'Save and continue').click();

            cy.get('.govuk-error-summary').within(() => {
                cy.contains('Enter first names');
                cy.contains('Enter last name');
            });

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        });

        it('errors when names too long', () => {
            cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
            cy.get('#f-last-name').invoke('val', 'b'.repeat(62));
            cy.get('#f-other-names').invoke('val', 'c'.repeat(51));

            cy.contains('button', 'Save and continue').click();

            cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
            cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
            cy.contains('[for=f-other-names] + div + .govuk-error-message', 'Other names you are known by must be 50 characters or less');
        });
    });

    describe('after completing', () => {
        beforeEach(() => {
            cy.visit('/fixtures?redirect=/your-name&progress=chooseYourAttorneys');
        });

        it('shows task list button', () => {
            cy.contains('a', 'Return to task list');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });

        it('warns when name shared with other actor', () => {
            cy.get('#f-first-names').clear().type('Jessie');
            cy.get('#f-last-name').clear().type('Jones');
            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-name');

            cy.contains('There is also an attorney called Jessie Jones. An attorney cannot also be the donor. By saving this section, you are confirming that these are two different people with the same name.');

            cy.contains('button', 'Save and continue').click();
            cy.url().should('contain', '/your-date-of-birth');
        });
    });
});
