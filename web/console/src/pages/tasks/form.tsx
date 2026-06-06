import React, { useState, useEffect } from 'react';
import { ApiClient } from '../../api/client';
import { JsonEditor } from '../../components/JsonEditor';

interface TaskFormProps {
  template?: any; // If provided, we are editing
  onSave: () => void;
  onCancel: () => void;
  onError: (err: any) => void;
}

export const TaskForm: React.FC<TaskFormProps> = ({ template, onSave, onCancel, onError }) => {
  const [name, setName] = useState(template?.Name || '');
  const [description, setDescription] = useState(template?.Description || '');
  const [taskType, setTaskType] = useState(template?.TaskType || 'STANDARD');
  const [targetRoleId, setTargetRoleId] = useState(template?.TargetRoleID || '');
  const [priority, setPriority] = useState(template?.Priority !== undefined ? template.Priority : 3);
  const [duration, setDuration] = useState(template?.EstimatedDurationMinutes !== undefined ? template.EstimatedDurationMinutes : 15);
  
  // Checklist template holds a JSON list of strings (steps)
  const [checklist, setChecklist] = useState<string[]>(() => {
    if (template?.ChecklistTemplate) {
      try {
        if (Array.isArray(template.ChecklistTemplate)) {
          return template.ChecklistTemplate;
        }
        return JSON.parse(template.ChecklistTemplate);
      } catch (err) {
        console.error("Failed parsing checklist template", err);
      }
    }
    return [];
  });

  const [newStep, setNewStep] = useState('');
  const [metadataStr, setMetadataStr] = useState(template?.Metadata ? JSON.stringify(template.Metadata, null, 2) : '{}');

  const [roles, setRoles] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [isMetadataValid, setIsMetadataValid] = useState(true);

  const getStepText = (step: any) => {
    if (typeof step === 'object' && step !== null) {
      return step.step || step.action || JSON.stringify(step);
    }
    return String(step);
  };

  useEffect(() => {
    ApiClient.fetchRoles()
      .then((data) => {
        setRoles(data || []);
      })
      .catch(onError);
  }, []);

  const handleAddStep = () => {
    if (!newStep.trim()) return;
    setChecklist([...checklist, newStep.trim()]);
    setNewStep('');
  };

  const handleRemoveStep = (index: number) => {
    const next = [...checklist];
    next.splice(index, 1);
    setChecklist(next);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name) {
      alert('Please fill out all required fields.');
      return;
    }

    let parsedMetadata = {};
    try {
      parsedMetadata = JSON.parse(metadataStr);
    } catch (err) {
      alert('Invalid JSON in Metadata field.');
      return;
    }

    const payload = {
      ...template,
      Name: name,
      Description: description,
      TaskType: taskType,
      TargetRoleID: targetRoleId || null,
      Priority: Number(priority),
      EstimatedDurationMinutes: Number(duration),
      ChecklistTemplate: checklist, // GORM expects jsonb array
      Metadata: parsedMetadata,
    };

    try {
      setSubmitting(true);
      if (template?.ID) {
        await ApiClient.updateTaskTemplate(template.ID, payload);
      } else {
        await ApiClient.createTaskTemplate(payload);
      }
      onSave();
    } catch (err) {
      onError(err);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="panel-card flex-1 flex flex-col min-h-0">
      <div className="panel-header">
        <h2 className="panel-title">{template ? 'Edit Workflow Template' : 'Design Operational Blueprint'}</h2>
        <p style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
          Create rigid, sequential instructions for compliance audits and team checkpoints.
        </p>
      </div>

      <div className="panel-body-scrollable flex-1 p-6">
        <form onSubmit={handleSubmit} className="flex flex-col gap-6 max-w-3xl">
          {/* Task Name */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Task Blueprint Name *</label>
            <input
              type="text"
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
              placeholder="e.g. Daily Freezer Temperature Audit"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          {/* Description */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Description / Purpose</label>
            <textarea
              className="site-meta-pill"
              style={{ borderRadius: '8px', padding: '12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)', minHeight: '80px', resize: 'vertical' }}
              placeholder="Describe the scope, tools, and compliance guidelines for this task..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Required Role */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Target Role Assignee</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={targetRoleId}
                onChange={(e) => setTargetRoleId(e.target.value)}
              >
                <option value="" style={{ background: 'var(--bg-main)' }}>All Personnel (No constraint)</option>
                {roles.map((r) => (
                  <option key={r.ID} value={r.ID} style={{ background: 'var(--bg-main)' }}>{r.Name}</option>
                ))}
              </select>
            </div>

            {/* Task Type */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Workflow Paradigm *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={taskType}
                onChange={(e) => setTaskType(e.target.value)}
                required
              >
                <option value="STANDARD" style={{ background: 'var(--bg-main)' }}>Standard Checkpoint (Human led)</option>
                <option value="AI_DRAFT" style={{ background: 'var(--bg-main)' }}>AI Co-Piloted (Grounded)</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Priority */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Priority Classification *</label>
              <select
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={priority}
                onChange={(e) => setPriority(Number(e.target.value))}
                required
              >
                <option value="1" style={{ background: 'var(--bg-main)' }}>Critical (Immediate safety/regulatory SLA)</option>
                <option value="2" style={{ background: 'var(--bg-main)' }}>High (Same-day completion)</option>
                <option value="3" style={{ background: 'var(--bg-main)' }}>Standard (Routine operational checks)</option>
              </select>
            </div>

            {/* SLA Duration */}
            <div className="flex flex-col gap-2">
              <label className="a2ui-label font-semibold" style={{ padding: 0 }}>SLA Estimated Duration (Minutes) *</label>
              <input
                type="number"
                min="1"
                className="site-meta-pill"
                style={{ borderRadius: '8px', padding: '10px 14px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)', color: 'var(--text-primary)' }}
                value={duration}
                onChange={(e) => setDuration(Number(e.target.value) || 15)}
                required
              />
            </div>
          </div>

          {/* Checklist Step Builder */}
          <div className="flex flex-col gap-2 p-4 rounded-xl" style={{ background: 'rgba(255,255,255,0.015)', border: '1px dashed var(--panel-border)' }}>
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Checklist Blueprint Steps ({checklist.length})</label>
            
            {/* Steps List */}
            {checklist.length > 0 ? (
              <div className="flex flex-col gap-2 mb-4">
                {checklist.map((step, idx) => (
                  <div key={idx} className="flex justify-between items-center p-3 rounded-lg border" style={{ background: 'var(--checklist-step-bg)', borderColor: 'var(--checklist-step-border)' }}>
                    <span className="text-sm" style={{ color: 'var(--text-secondary)' }}>{idx + 1}. {getStepText(step)}</span>
                    <button
                      type="button"
                      className="a2ui-btn-action text-xs"
                      style={{ borderColor: 'var(--priority-critical)', color: 'var(--priority-critical)', padding: '2px 8px' }}
                      onClick={() => handleRemoveStep(idx)}
                    >
                      Remove
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted mb-4">No checklist steps defined. Add a step below.</p>
            )}

            {/* Input Row */}
            <div className="flex gap-2">
              <input
                type="text"
                className="site-meta-pill flex-1"
                style={{ borderRadius: '8px', padding: '8px 12px', background: 'var(--input-bg)', border: '1px solid var(--panel-border)' }}
                placeholder="Type a sequential checkpoint instruction..."
                value={newStep}
                onChange={(e) => setNewStep(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAddStep(); } }}
              />
              <button
                type="button"
                className="btn-primary"
                style={{ padding: '8px 16px' }}
                onClick={handleAddStep}
              >
                + Add Step
              </button>
            </div>
          </div>

          {/* Metadata */}
          <div className="flex flex-col gap-2">
            <label className="a2ui-label font-semibold" style={{ padding: 0 }}>Metadata (JSONB) *</label>
            <JsonEditor
              value={metadataStr}
              onChange={(val) => setMetadataStr(val)}
              onValidate={(valid) => setIsMetadataValid(valid)}
              placeholder="{}"
            />
          </div>

          {/* Actions */}
          <div className="flex gap-4 mt-4">
            <button
              type="submit"
              className="btn-primary"
              disabled={submitting || !isMetadataValid}
              style={{ padding: '10px 24px' }}
            >
              {submitting ? 'Saving...' : 'Save Template'}
            </button>
            <button
              type="button"
              className="a2ui-btn-action"
              onClick={onCancel}
              style={{ padding: '10px 24px' }}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
