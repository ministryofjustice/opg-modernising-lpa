describe('LPA progress', () => {
    it('shows the progress of the LPA', () => {
        cy.visit('/testing-start?redirect=/progress');
        cy.contains('li', 'LPA signed In progress');
        cy.contains('li', 'Certificate provider has made their declaration Not started');
        cy.contains('li', 'Attorneys have made their declaration Not started');
        cy.contains('li', 'LPA submitted to the OPG Not started');
        cy.contains('li', 'Statutory waiting period Not started');
        cy.contains('li', 'LPA registered Not started');

        cy.visit('/testing-start?redirect=/progress&completeLpa=1');
        cy.contains('li', 'LPA signed Completed');
        cy.contains('li', 'Certificate provider has made their declaration In progress');
        cy.contains('li', 'Attorneys have made their declaration Not started');
        cy.contains('li', 'LPA submitted to the OPG Not started');
        cy.contains('li', 'Statutory waiting period Not started');
        cy.contains('li', 'LPA registered Not started');
    })
});
