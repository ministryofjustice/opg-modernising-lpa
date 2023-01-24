describe('Dashboard', () => {
    beforeEach(() => {
        cy.visit('/testing-start?redirect=/your-details');
        cy.get('#f-first-names').type('John');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('1');
        cy.get('#f-date-of-birth-month').type('2');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/lpa-type');
        cy.get('#f-lpa-type').check();
        cy.contains('button', 'Continue').click();

        cy.visit('/dashboard');
    });

    it('shows my lasting power of attorney', () => {        
        cy.contains('Property and affairs');
        cy.contains('John Doe');
        cy.contains('a', 'Continue').click();

        cy.url().should('contain', '/task-list');
    });

    it('can create another', () => {
        cy.visit('/dashboard');

        cy.contains('button', 'Create another LPA').click();

        cy.get('#f-first-names').type('Jane');
        cy.get('#f-last-name').type('Doe');
        cy.get('#f-date-of-birth').type('2');
        cy.get('#f-date-of-birth-month').type('3');
        cy.get('#f-date-of-birth-year').type('1990');
        cy.contains('button', 'Continue').click();

        cy.visitLpa('/lpa-type');
        cy.get('#f-lpa-type-2').check();
        cy.contains('button', 'Continue').click();

        cy.visit('/dashboard');

        cy.contains('Property and affairs: John Doe');
        cy.contains('Personal welfare: Jane Doe');
    });
});
