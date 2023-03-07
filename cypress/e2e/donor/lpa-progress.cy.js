describe('LPA progress', () => {
    it('when nothing completed', () => {
        cy.visit('/testing-start?redirect=/progress');

        cy.checkA11yApp();

        cy.contains('li', 'LPA signed In progress');
        cy.contains('li', 'Certificate provider has made their declaration Not started');
        cy.contains('li', 'Attorneys have made their declaration Not started');
        cy.contains('li', 'LPA submitted to the OPG Not started');
        cy.contains('li', 'Statutory waiting period Not started');
        cy.contains('li', 'LPA registered Not started');
    })

    it('when LPA submitted', () => {
        cy.visit('/testing-start?redirect=/progress&completeLpa=1');

        cy.checkA11yApp();

        cy.contains('li', 'LPA signed Completed');
        cy.contains('li', 'Certificate provider has made their declaration In progress');
        cy.contains('li', 'Attorneys have made their declaration Not started');
        cy.contains('li', 'LPA submitted to the OPG Not started');
        cy.contains('li', 'Statutory waiting period Not started');
        cy.contains('li', 'LPA registered Not started');
    })

    it('when certificate provided', () => {
        cy.visit('/testing-start?redirect=/progress&completeLpa=1&provideCertificate=1&asDonor=1');

        cy.checkA11yApp();

        cy.contains('li', 'LPA signed Completed');
        cy.contains('li', 'Certificate provider has made their declaration Completed');
        cy.contains('li', 'Attorneys have made their declaration In progress');
        cy.contains('li', 'LPA submitted to the OPG Not started');
        cy.contains('li', 'Statutory waiting period Not started');
        cy.contains('li', 'LPA registered Not started');
    })
});
