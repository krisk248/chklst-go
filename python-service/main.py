"""
Python Microservice for Excel/PDF Generation
Embedded within chklst-go binary
"""
from fastapi import FastAPI, HTTPException
from fastapi.responses import Response
from pydantic import BaseModel
from typing import List, Dict, Any, Optional
from datetime import datetime
import uvicorn

from services.excel_service import ExcelService
from services.pdf_service import PDFService

# Initialize FastAPI app
app = FastAPI(
    title="chklst-go Reports Service",
    description="Python microservice for Excel and PDF report generation",
    version="1.0.0"
)

# Initialize services
excel_service = ExcelService()
pdf_service = PDFService()


# Request Models
class DeploymentData(BaseModel):
    """Deployment data model"""
    id: int
    jira_id: str
    project_name: str
    component_name: str
    developer: str
    build_server: str
    deploy_server: str
    environment: str
    timestamp: str
    status: Optional[str] = None
    notes: Optional[str] = None


class ExcelReportRequest(BaseModel):
    """Request model for Excel report generation"""
    deployments: List[Dict[str, Any]]
    month: int
    year: int


class PDFReportRequest(BaseModel):
    """Request model for PDF report generation"""
    deployments: List[Dict[str, Any]]
    month: int
    year: int


class StatsReportRequest(BaseModel):
    """Request model for statistics report"""
    stats: Dict[str, Any]
    month: int
    year: int


# Health Check
@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "reports-service",
        "version": "1.0.0",
        "timestamp": datetime.now().isoformat()
    }


# Excel Endpoints
@app.post("/api/reports/excel/deployments")
async def generate_excel_deployments(request: ExcelReportRequest):
    """
    Generate Excel report for deployments

    Args:
        request: Excel report request with deployments data

    Returns:
        Excel file as binary response
    """
    try:
        excel_bytes = excel_service.generate_deployment_report(
            deployments=request.deployments,
            month=request.month,
            year=request.year
        )

        filename = f"deployments_{request.year}_{request.month:02d}.xlsx"

        return Response(
            content=excel_bytes,
            media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            headers={
                "Content-Disposition": f"attachment; filename={filename}"
            }
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to generate Excel report: {str(e)}")


@app.post("/api/reports/excel/statistics")
async def generate_excel_statistics(request: StatsReportRequest):
    """
    Generate Excel report for statistics

    Args:
        request: Statistics report request

    Returns:
        Excel file as binary response
    """
    try:
        excel_bytes = excel_service.generate_statistics_report(
            stats=request.stats,
            month=request.month,
            year=request.year
        )

        filename = f"statistics_{request.year}_{request.month:02d}.xlsx"

        return Response(
            content=excel_bytes,
            media_type="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
            headers={
                "Content-Disposition": f"attachment; filename={filename}"
            }
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to generate Excel statistics: {str(e)}")


# PDF Endpoints
@app.post("/api/reports/pdf/deployments")
async def generate_pdf_deployments(request: PDFReportRequest):
    """
    Generate PDF report for deployments

    Args:
        request: PDF report request with deployments data

    Returns:
        PDF file as binary response
    """
    try:
        pdf_bytes = pdf_service.generate_deployment_report(
            deployments=request.deployments,
            month=request.month,
            year=request.year
        )

        filename = f"deployments_{request.year}_{request.month:02d}.pdf"

        return Response(
            content=pdf_bytes,
            media_type="application/pdf",
            headers={
                "Content-Disposition": f"attachment; filename={filename}"
            }
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to generate PDF report: {str(e)}")


@app.post("/api/reports/pdf/statistics")
async def generate_pdf_statistics(request: StatsReportRequest):
    """
    Generate PDF report for statistics

    Args:
        request: Statistics report request

    Returns:
        PDF file as binary response
    """
    try:
        pdf_bytes = pdf_service.generate_statistics_report(
            stats=request.stats,
            month=request.month,
            year=request.year
        )

        filename = f"statistics_{request.year}_{request.month:02d}.pdf"

        return Response(
            content=pdf_bytes,
            media_type="application/pdf",
            headers={
                "Content-Disposition": f"attachment; filename={filename}"
            }
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to generate PDF statistics: {str(e)}")


# Root endpoint
@app.get("/")
async def root():
    """Root endpoint"""
    return {
        "service": "chklst-go Reports Service",
        "version": "1.0.0",
        "endpoints": {
            "health": "/health",
            "excel_deployments": "/api/reports/excel/deployments",
            "excel_statistics": "/api/reports/excel/statistics",
            "pdf_deployments": "/api/reports/pdf/deployments",
            "pdf_statistics": "/api/reports/pdf/statistics"
        }
    }


if __name__ == "__main__":
    # Run with uvicorn
    uvicorn.run(
        "main:app",
        host="127.0.0.1",
        port=8001,
        log_level="info",
        reload=False  # Disable reload for production
    )
