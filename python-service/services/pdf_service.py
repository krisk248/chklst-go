"""
PDF report generation service using ReportLab and Matplotlib
"""
from reportlab.lib import colors
from reportlab.lib.pagesizes import letter, A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.platypus import (
    SimpleDocTemplate, Table, TableStyle, Paragraph,
    Spacer, PageBreak, Image
)
from reportlab.lib.enums import TA_CENTER, TA_LEFT
from datetime import datetime
from typing import List, Dict, Any
import io
import matplotlib.pyplot as plt
import matplotlib
matplotlib.use('Agg')  # Non-interactive backend


class PDFService:
    """Service for generating PDF reports"""

    def __init__(self):
        self.styles = getSampleStyleSheet()
        self.title_style = ParagraphStyle(
            'CustomTitle',
            parent=self.styles['Heading1'],
            fontSize=18,
            textColor=colors.HexColor('#366092'),
            spaceAfter=30,
            alignment=TA_CENTER
        )
        self.heading_style = ParagraphStyle(
            'CustomHeading',
            parent=self.styles['Heading2'],
            fontSize=14,
            textColor=colors.HexColor('#366092'),
            spaceAfter=12
        )

    def generate_deployment_report(
        self,
        deployments: List[Dict[str, Any]],
        month: int,
        year: int
    ) -> bytes:
        """
        Generate PDF report for deployments

        Args:
            deployments: List of deployment dictionaries
            month: Report month (1-12)
            year: Report year

        Returns:
            PDF file as bytes
        """
        buffer = io.BytesIO()
        doc = SimpleDocTemplate(
            buffer,
            pagesize=letter,
            rightMargin=30,
            leftMargin=30,
            topMargin=30,
            bottomMargin=30
        )

        # Container for the 'Flowable' objects
        elements = []

        # Title
        title_text = f"Deployment Report - {datetime(year, month, 1).strftime('%B %Y')}"
        elements.append(Paragraph(title_text, self.title_style))
        elements.append(Spacer(1, 12))

        # Metadata
        metadata_text = f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
        elements.append(Paragraph(metadata_text, self.styles['Normal']))
        elements.append(Spacer(1, 20))

        # Summary Statistics
        elements.append(Paragraph("Summary", self.heading_style))

        summary_data = [
            ["Total Deployments", str(len(deployments))],
            ["Report Period", datetime(year, month, 1).strftime('%B %Y')],
        ]

        # Calculate environment breakdown
        if deployments:
            env_counts = {}
            for dep in deployments:
                env = dep.get('environment', 'Unknown')
                env_counts[env] = env_counts.get(env, 0) + 1

            summary_data.append(["", ""])
            summary_data.append(["Environment Breakdown", ""])
            for env, count in sorted(env_counts.items()):
                summary_data.append([f"  {env}", str(count)])

        summary_table = Table(summary_data, colWidths=[3*inch, 2*inch])
        summary_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#366092')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 11),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 12),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
        ]))

        elements.append(summary_table)
        elements.append(Spacer(1, 30))

        # Deployments Table
        if deployments:
            elements.append(Paragraph("Deployment Details", self.heading_style))
            elements.append(Spacer(1, 12))

            # Table headers
            table_data = [[
                'Jira ID',
                'Project',
                'Component',
                'Environment',
                'Developer',
                'Date'
            ]]

            # Table rows
            for dep in deployments[:50]:  # Limit to first 50 to avoid huge PDFs
                table_data.append([
                    dep.get('jira_id', ''),
                    dep.get('project_name', '')[:20],  # Truncate long names
                    dep.get('component_name', '')[:20],
                    dep.get('environment', ''),
                    dep.get('developer', '')[:15],
                    datetime.fromisoformat(dep['timestamp']).strftime('%m/%d')
                    if 'timestamp' in dep else ''
                ])

            # Create table
            deployment_table = Table(
                table_data,
                colWidths=[0.8*inch, 1.3*inch, 1.3*inch, 1*inch, 1*inch, 0.7*inch]
            )

            deployment_table.setStyle(TableStyle([
                # Header row styling
                ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#366092')),
                ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
                ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
                ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
                ('FONTSIZE', (0, 0), (-1, 0), 9),
                ('BOTTOMPADDING', (0, 0), (-1, 0), 8),

                # Data rows styling
                ('FONTNAME', (0, 1), (-1, -1), 'Helvetica'),
                ('FONTSIZE', (0, 1), (-1, -1), 8),
                ('ROWBACKGROUNDS', (0, 1), (-1, -1), [colors.white, colors.HexColor('#f0f0f0')]),

                # Grid
                ('GRID', (0, 0), (-1, -1), 0.5, colors.grey),
            ]))

            elements.append(deployment_table)

            if len(deployments) > 50:
                elements.append(Spacer(1, 12))
                note = f"Note: Showing first 50 of {len(deployments)} deployments"
                elements.append(Paragraph(note, self.styles['Italic']))

        else:
            elements.append(Paragraph("No deployments found for this period.", self.styles['Normal']))

        # Add chart if we have data
        if deployments:
            elements.append(PageBreak())
            elements.append(Paragraph("Deployment Trends", self.heading_style))
            elements.append(Spacer(1, 12))

            # Generate chart
            chart_buffer = self._generate_deployment_chart(deployments, month, year)
            if chart_buffer:
                img = Image(chart_buffer, width=6*inch, height=4*inch)
                elements.append(img)

        # Build PDF
        doc.build(elements)

        buffer.seek(0)
        return buffer.getvalue()

    def _generate_deployment_chart(
        self,
        deployments: List[Dict[str, Any]],
        month: int,
        year: int
    ) -> io.BytesIO:
        """Generate deployment trend chart"""
        try:
            # Count deployments by environment
            env_counts = {}
            for dep in deployments:
                env = dep.get('environment', 'Unknown')
                env_counts[env] = env_counts.get(env, 0) + 1

            # Create pie chart
            fig, ax = plt.subplots(figsize=(8, 6))

            colors_list = ['#366092', '#4a7ba7', '#6496c8', '#7eb0db', '#98caee']
            ax.pie(
                env_counts.values(),
                labels=env_counts.keys(),
                autopct='%1.1f%%',
                startangle=90,
                colors=colors_list[:len(env_counts)]
            )

            ax.set_title(
                f'Deployments by Environment - {datetime(year, month, 1).strftime("%B %Y")}',
                fontsize=14,
                fontweight='bold'
            )

            # Save to buffer
            chart_buffer = io.BytesIO()
            plt.savefig(chart_buffer, format='png', dpi=150, bbox_inches='tight')
            plt.close(fig)

            chart_buffer.seek(0)
            return chart_buffer

        except Exception as e:
            print(f"Error generating chart: {e}")
            return None

    def generate_statistics_report(
        self,
        stats: Dict[str, Any],
        month: int,
        year: int
    ) -> bytes:
        """
        Generate PDF report for statistics

        Args:
            stats: Statistics dictionary
            month: Report month
            year: Report year

        Returns:
            PDF file as bytes
        """
        buffer = io.BytesIO()
        doc = SimpleDocTemplate(buffer, pagesize=letter)

        elements = []

        # Title
        title = f"Deployment Statistics - {datetime(year, month, 1).strftime('%B %Y')}"
        elements.append(Paragraph(title, self.title_style))
        elements.append(Spacer(1, 20))

        # Statistics table
        table_data = [["Metric", "Value"]]
        for key, value in stats.items():
            table_data.append([key.replace('_', ' ').title(), str(value)])

        stats_table = Table(table_data, colWidths=[4*inch, 2*inch])
        stats_table.setStyle(TableStyle([
            ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#366092')),
            ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
            ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
            ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
            ('FONTSIZE', (0, 0), (-1, 0), 12),
            ('BOTTOMPADDING', (0, 0), (-1, 0), 12),
            ('GRID', (0, 0), (-1, -1), 1, colors.grey),
            ('ROWBACKGROUNDS', (0, 1), (-1, -1), [colors.white, colors.HexColor('#f0f0f0')]),
        ]))

        elements.append(stats_table)

        doc.build(elements)
        buffer.seek(0)
        return buffer.getvalue()
