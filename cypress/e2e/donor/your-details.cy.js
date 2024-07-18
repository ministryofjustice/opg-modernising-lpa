describe('Donor details', () => {
    beforeEach(() => {
        cy.visit('/fixtures?redirect=/your-details');
    });

    it('can be submitted', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.get('#f-can-sign').check({ force: true });

        cy.checkA11yApp();

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-address');
    });

    it('errors when empty', () => {
        cy.contains('button', 'Continue').click();

        cy.get('.govuk-error-summary').within(() => {
            cy.contains('Enter first names');
            cy.contains('Enter last name');
            cy.contains('Enter date of birth');
        });

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'Enter first names');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Enter last name');
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Enter date of birth');
    });

    it('errors when names too long', () => {
        cy.get('#f-first-names').invoke('val', 'a'.repeat(54));
        cy.get('#f-last-name').invoke('val', 'b'.repeat(62));
        cy.get('#f-other-names').invoke('val', 'c'.repeat(51));

        cy.contains('button', 'Continue').click();

        cy.contains('[for=f-first-names] + div + .govuk-error-message', 'First names must be 53 characters or less');
        cy.contains('[for=f-last-name] + .govuk-error-message', 'Last name must be 61 characters or less');
        cy.contains('[for=f-other-names] + div + .govuk-error-message', 'Other names you are known by must be 50 characters or less');
    });

    it('errors when invalid dates of birth', () => {
        cy.get('#f-date-of-birth').type('1');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must include a month and year');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('2222');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be in the past');

        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').clear().type('1990');
        cy.contains('button', 'Continue').click();
        cy.contains('#date-of-birth-hint + .govuk-error-message', 'Date of birth must be a real date');
    });

    it('warns when name shared with other actor', () => {
        cy.visit('/fixtures?redirect=/your-details&progress=chooseYourAttorneys');

        cy.get('#f-first-names').clear().type('Jessie');
        cy.get('#f-last-name').clear().type('Jones');
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-details');

        cy.contains('There is also an attorney called Jessie Jones. An attorney cannot also be the donor. By saving this section, you are confirming that these are two different people with the same name.');

        cy.get('#f-can-sign').check();
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-address');
    });

    it('permanently warns when date of birth is under 18', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type(new Date().getFullYear() - 1);
        cy.get('#f-can-sign').check({ force: true });
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-details');

        cy.contains('You are under 18. By continuing, you understand that you must be at least 18 years old on the date you sign the LPA, or it will be rejected.');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-address');

        cy.visitLpa("/task-list")
        cy.contains('a', 'Provide your details').click()
        cy.url().should('contain', '/your-details');

        cy.contains('You are under 18. By continuing, you understand that you must be at least 18 years old on the date you sign the LPA, or it will be rejected.');
    });

    it('warns when date of birth is over 100', () => {
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1900');
        cy.get('#f-can-sign').check({ force: true });
        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-details');

        cy.contains('By continuing, you confirm that this person is more than 100 years old. If not, please change their date of birth.');

        cy.contains('button', 'Continue').click();
        cy.url().should('contain', '/your-address');
    });
});
