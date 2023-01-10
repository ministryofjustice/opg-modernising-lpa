describe('Confirm your identity and sign', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/your-details&withIncompleteAttorneys=1&withCP=1&withPayment=1');
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();
        cy.visit('/task-list');
    });

    it('can be completed', () => {
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('label', 'Your GOV.UK One Login Identity').click();
        cy.contains('button', 'Continue').click();

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Your GOV.UK One Login Identity');

        // can't click continue as the real flow would begin
        cy.visit('/read-your-lpa');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.url().should('contain', '/read-your-lpa');

        cy.contains('h2', "LPA decisions");
        cy.contains('h2', "People named on the LPA");
        cy.contains('h3', "Donor");
        cy.contains('h3', "Attorneys");
        cy.contains('h3', "Replacement attorney");

        cy.contains('a', 'Continue to signing page').click();

        cy.url().should('contain', '/sign-your-lpa');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('h1', "Sign your LPA");

        cy.contains('label', 'is witnessing me sign this LPA').click();
        cy.contains('label', 'I want to apply to register this LPA').click();
        cy.contains('button', 'Submit my signature').click();

        cy.url().should('contain', '/witnessing-your-signature');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/witnessing-as-certificate-provider');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('h1', "Witnessing as the certificate provider");

        cy.get('#f-witness-code').type('1234');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/you-have-submitted-your-lpa');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('h1', "You've submitted your LPA");

        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/dashboard');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });
    });

    it('can be restarted', () => {
        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'Not started')
            .find('a')
            .click();

        cy.contains('label', 'Your GOV.UK One Login Identity').click();
        cy.contains('button', 'Continue').click();

        cy.visit('/task-list');

        cy.contains('li', "Confirm your identity and sign")
            .should('contain', 'In progress')
            .find('a')
            .click();

        cy.contains('button', 'Continue').click();
        cy.contains('Your GOV.UK One Login Identity');
    });
});
