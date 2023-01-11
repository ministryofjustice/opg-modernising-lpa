describe('Choose replacement attorneys summary', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/choose-replacement-attorneys-summary&withIncompleteAttorneys=1&cookiesAccepted=1');
    });

    it('multiple attorneys details are listed', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('You have added 2 replacement attorneys');

        cy.contains('John Smith');
        cy.contains('2 January 2000');
        cy.contains('2 RICHMOND PLACE');
        cy.contains('B14 7ED');

        cy.contains('Joan Smith');
        cy.contains('2 January 1998');

        cy.visit('/task-list')
        cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('In progress (2)')
    });

    it('can amend attorney details', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#name-1').contains('a', 'Change').click();

        cy.url().should('contain', '/choose-replacement-attorneys');
        cy.url().should('contain', 'from=%2fchoose-replacement-attorneys-summary');
        cy.url().should('match', /id=\w*/);

        cy.get('#f-first-names').clear().type('Mark');

        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-summary');

        cy.contains('Mark Smith');
    });

    it('can amend attorney address', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#address-2').contains('a', 'Change').click();

        cy.url().should('contain', '/choose-replacement-attorneys-address');
        cy.url().should('contain', 'from=%2fchoose-replacement-attorneys-summary');
        cy.url().should('match', /id=\w*/);

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.get('#f-select-address').select('4 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-address');
        cy.get('#f-address-line-1').should('have.value', '4 RICHMOND PLACE');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-summary');

        cy.contains('dd', '4 RICHMOND PLACE');

        cy.visit('/task-list')
        cy.contains('a', 'Choose your replacement attorneys').parent().parent().contains('Completed (2)')
    });

    it('can add another attorney from summary page', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#f-add-attorney').check('yes');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys');

        cy.get('#f-first-names').clear().type('Bob Arnold');
        cy.get('#f-last-name').clear().type('Jones');
        cy.get('#f-email').clear().type('dd@example.org');
        cy.get('input[name="date-of-birth-day"]').clear().type('31');
        cy.get('input[name="date-of-birth-month"]').clear().type('12');
        cy.get('input[name="date-of-birth-year"]').clear().type('1995');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-address');

        cy.get('#f-lookup-postcode').type('B14 7ED');
        cy.contains('button', 'Find address').click();

        cy.get('#f-select-address').select('5 RICHMOND PLACE, BIRMINGHAM, B14 7ED');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-address');
        cy.get('#f-address-line-1').should('have.value', '5 RICHMOND PLACE');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-summary');

        cy.contains('Bob Arnold Jones');
        cy.contains('31 December 1995');
        cy.contains('5 RICHMOND PLACE');
        cy.contains('B14 7ED');
    });

    it('can remove an attorney', () => {
        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('#remove-attorney-1').contains('a', 'Remove').click();

        cy.url().should('contain', '/remove-replacement-attorney');
        cy.url().should('match', /id=\w*/);

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.contains('Are you sure you want to remove John Smith?');

        cy.get('#f-remove-attorney').check('yes');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/choose-replacement-attorneys-summary');

        cy.injectAxe();
        cy.checkA11y(null, { rules: { region: { enabled: false } } });

        cy.get('main').should('not.contain', 'John Smith');

        cy.get('#remove-attorney-1').contains('a', 'Remove').click();
        cy.get('#f-remove-attorney').check('yes');
        cy.contains('button', 'Continue').click();

        cy.url().should('contain', '/do-you-want-replacement-attorneys');
    });
});
