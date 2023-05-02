describe('Dashboard', () => {
    context('with incomplete LPA', () => {
        beforeEach(() => {
            cy.visit('/testing-start?redirect=/dashboard&withDonorDetails=1');
        });

        it('shows my lasting power of attorney', () => {
            cy.contains('Property and affairs');
            cy.contains('Jamie Smith');
            cy.contains('a', 'Continue').click();

            cy.url().should('contain', '/task-list');
        });

        it('can create another', () => {
            cy.contains('button', 'Create another LPA').click();

            cy.get('#f-first-names').type('Jane');
            cy.get('#f-last-name').type('Smith');
            cy.get('#f-date-of-birth').type('2');
            cy.get('#f-date-of-birth-month').type('3');
            cy.get('#f-date-of-birth-year').type('1990');
            cy.contains('button', 'Continue').click();

            cy.visitLpa('/lpa-type');

            cy.get('#f-lpa-type-2').check();
            cy.contains('button', 'Continue').click();

            cy.visit('/dashboard');

            cy.contains('Property and affairs: Jamie Smith');
            cy.contains('Personal welfare: Jane Smith');
        });
    })

    context('with completed LPA', () => {
        it('completed LPAs have a track progress button', () => {
            cy.visit('/testing-start?redirect=/dashboard&completeLpa=1')

            cy.get('button').should('not.contain', 'Continue');

            cy.contains('Property and affairs');
            cy.contains('Jamie Smith');
            cy.contains('a', 'Track LPA progress').click();

            cy.url().should('contain', '/progress');
        });
    })
});
